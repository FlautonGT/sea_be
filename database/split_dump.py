import re
import os
import csv
import io

INPUT_FILE = r'c:\Users\USER\Website Pribadi\Gate V2\Backend\database\db_converted.sql'
OUTPUT_DIR = r'c:\Users\USER\Website Pribadi\Gate V2\Backend\database\migrations'

if not os.path.exists(OUTPUT_DIR):
    os.makedirs(OUTPUT_DIR)

def read_file():
    with open(INPUT_FILE, 'r', encoding='utf-8') as f:
        return f.read()

def parse_sql(content):
    # Split by "-- Name: " markers usually used by pg_dump
    # But wait, pg_dump uses "-- Name: ...; Type: ...; Schema: ...; Owner: ..."
    # And there are also "COPY ..." blocks which might not have a preceding marker block immediately if data follows index?
    # Actually checking the file:
    # -- Name: ...
    # CREATE TABLE ...
    # ...
    # -- Data for Name: ...
    # COPY ...
    
    # We can split by "-- Name: " but we need to keep the content.
    # Regex to find blocks.
    
    # Alternative: iterate line by line to build state.
    
    blocks = []
    current_block = []
    current_type = None
    current_name = None
    
    lines = content.splitlines()
    i = 0
    while i < len(lines):
        line = lines[i]
        
        if line.startswith("-- Name:"):
            # Start of a new relevant block?
            # Save previous block if exists
            if current_block:
                blocks.append({
                    'type': current_type,
                    'name': current_name,
                    'content': "\n".join(current_block)
                })
                current_block = []
                current_type = None
                current_name = None
            
            # Parse header
            # -- Name: admins; Type: TABLE; Schema: public; Owner: seaply
            match = re.search(r'-- Name: (.+?); Type: (.+?);', line)
            if match:
                current_name = match.group(1).strip()
                current_type = match.group(2).strip()
            
            current_block.append(line)
        
        elif line.startswith("-- Data for Name:"):
             # Data block
            if current_block:
                blocks.append({
                    'type': current_type,
                    'name': current_name,
                    'content': "\n".join(current_block)
                })
                current_block = []
            
            match = re.search(r'-- Data for Name: (.+?); Type: (.+?);', line)
            if match:
                current_name = match.group(1).strip()
                current_type = match.group(2).strip() # TABLE DATA
            current_block.append(line)
        
        else:
            current_block.append(line)
            
        i += 1
        
    if current_block:
        blocks.append({
            'type': current_type,
            'name': current_name,
            'content': "\n".join(current_block)
        })
        
    return blocks

def escape_sql_string(val):
    if val == '\\N':
        return 'NULL'
    # Escape single quotes
    val = val.replace("'", "''")
    return f"'{val}'"

def convert_copy_to_insert(copy_block):
    # Extract COPY statement
    lines = copy_block.splitlines()
    copy_cmd = None
    data_lines = []
    
    for line in lines:
        if line.startswith("COPY"):
            copy_cmd = line
        elif line.startswith(r"\.") or line.strip() == "":
            continue
        elif not line.startswith("--"):
            data_lines.append(line)
            
    if not copy_cmd:
        return ""
        
    # COPY public.table (col1, col2) FROM stdin;
    match = re.match(r'COPY (.+?) \((.+?)\) FROM stdin;', copy_cmd)
    if not match:
        return "" # Could not parse COPY
        
    table_name = match.group(1)
    columns = match.group(2)
    
    inserts = []
    # Parse data lines
    # Usually tab separated
    for row_str in data_lines:
        # Need to handle potential escaped chars in TSV?
        # pg_dump text format: https://www.postgresql.org/docs/current/sql-copy.html
        # Backslash characters are \b, \f, \n, \r, \t, \v, \\, \digits
        # We can try rudimentary split by \t
        parts = row_str.split('\t')
        
        vals = [escape_sql_string(p) for p in parts]
        inserts.append(f"({', '.join(vals)})")
    
    if not inserts:
        return ""

    # Batch inserts to avoid huge lines? Or just one big statement?
    # One big statement is faster but might hit limits.
    # Group by 1000
    chunk_size = 1000
    sql_statements = []
    
    for k in range(0, len(inserts), chunk_size):
        chunk = inserts[k:k+chunk_size]
        sql = f"INSERT INTO {table_name} ({columns}) VALUES\n" + ",\n".join(chunk) + ";"
        sql_statements.append(sql)
        
    return "\n\n".join(sql_statements)

def main():
    content = read_file()
    
    # Pre-process content to move header blocks (before first -- Name:) into an Init block
    lines = content.splitlines()
    first_marker = -1
    for idx, line in enumerate(lines):
        if line.startswith("-- Name:"):
            first_marker = idx
            break
            
    init_content = ""
    if first_marker > 0:
        init_content = "\n".join(lines[:first_marker])
        content_rest = "\n".join(lines[first_marker:])
    else:
        content_rest = content

    blocks = parse_sql(content_rest)
    
    migrations = []
    
    # 0. Global Init
    migrations.append({
        'name': 'init_setup',
        'up': init_content,
        'down': '' # Hard to revert globals automatically
    })
    
    # Group By Table
    tables = {} # name -> {create: [], seeds: [], indexes: [], constraints: [], triggers: []}
    global_setup = [] # Extensions, Types, Functions
    fk_constraints = []
    
    for block in blocks:
        btype = block.get('type')
        bname = block.get('name')
        bcontent = block.get('content')
        
        if btype == 'TABLE':
            if bname not in tables:
                tables[bname] = {'create': [], 'seeds': [], 'indexes': [], 'constraints': [], 'triggers': []}
            tables[bname]['create'].append(bcontent)
        elif btype == 'TABLE DATA':
            # name is usually table name
            if bname in tables:
                insert_sql = convert_copy_to_insert(bcontent)
                tables[bname]['seeds'].append(insert_sql)
        elif btype == 'INDEX':
            # Name matches index name. Need to find table?
            # content usually: CREATE INDEX idx_name ON public.table ...
            match = re.search(r'ON public\.(\w+)', bcontent)
            if match:
                tname = match.group(1)
                if tname in tables:
                    tables[tname]['indexes'].append(bcontent)
        elif btype == 'TRIGGER':
            match = re.search(r'ON public\.(\w+)', bcontent)
            if match:
                tname = match.group(1)
                if tname in tables:
                    tables[tname]['triggers'].append(bcontent)
        elif btype == 'CONSTRAINT' or btype == 'FK CONSTRAINT':
             # Primary keys and Foreign keys
             # Content: ALTER TABLE ONLY public.table ADD CONSTRAINT ...
             match = re.search(r'ALTER TABLE ONLY public\.(\w+)', bcontent)
             if match:
                tname = match.group(1)
                if 'FOREIGN KEY' in bcontent:
                    fk_constraints.append(bcontent)
                else:
                    # PK or Unique
                    if tname in tables:
                        tables[tname]['constraints'].append(bcontent)
        elif btype in ['EXTENSION', 'TYPE', 'FUNCTION']:
            global_setup.append(bcontent)
        else:
            # Other stuff?
            if bcontent.strip():
                global_setup.append(bcontent)

    # 1. Update init with global setup
    migrations[0]['up'] += "\n\n".join(global_setup)
    
    # 2. Table Migrations
    # We want to maintain order. Dict preservation in Python 3.7+ used?
    # Iterate keys of tables.
    table_names = sorted(tables.keys()) # Or just iteration if insertion order matters?
    # In db_converted.sql, they seemed alphabetical.
    # It shouldn't matter for tables if FKs are at the end.
    
    for tname in table_names:
        tdata = tables[tname]
        up_sql = []
        up_sql.extend(tdata['create'])
        up_sql.extend(tdata['constraints']) # PKs and Uniques
        up_sql.extend(tdata['indexes'])
        up_sql.extend(tdata['triggers'])
        
        # Then Seeds
        if tdata['seeds']:
             up_sql.append("\n-- SEED DATA --\n")
             up_sql.extend(tdata['seeds'])
             
        migrations.append({
            'name': f"create_{tname}",
            'up': "\n\n".join(up_sql),
            'down': f"DROP TABLE IF EXISTS public.{tname};"
        })
        
    # 3. Foreign Keys
    if fk_constraints:
        migrations.append({
            'name': 'add_foreign_keys',
            'up': "\n\n".join(fk_constraints),
            'down': " -- TODO: Drop constraints if needed, or just dropping tables handles it"
        })

    # Write files
    seq = 1
    for mig in migrations:
        seq_str = f"{seq:06d}"
        slug = mig['name']
        up_name = f"{seq_str}_{slug}.up.sql"
        down_name = f"{seq_str}_{slug}.down.sql"
        
        with open(os.path.join(OUTPUT_DIR, up_name), 'w', encoding='utf-8') as f:
            f.write(mig['up'])
            
        with open(os.path.join(OUTPUT_DIR, down_name), 'w', encoding='utf-8') as f:
            f.write(mig['down'])
            
        print(f"Created {up_name} and {down_name}")
        seq += 1

if __name__ == "__main__":
    main()
