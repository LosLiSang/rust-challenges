#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""Redis RDB文件解析器 - 精确提取Key和Value"""

import struct

# 字节流数据
data = bytes([109,121,115,113,108,48,48,48,57,250,9,114,101,100,105,115,45,118,101,114,5,54,46,50,46,54,250,10,114,101,100,105,115,45,98,105,116,115,192,64,250,5,99,116,105,109,101,194,247,33,235,104,250,8,117,115,101,100,45,109,101,109,194,120,96,12,0,250,12,97,111,102,45,112,114,101,97,109,98,108,101,192,0,254,1,251,2,0,0,16,99,117,114,108,121,95,112,111,110,100,95,99,111,117,110,116,193,217,3,0,16,98,114,111,107,101,110,95,104,97,116,95,99,111,117,110,116,193,79,2,254,2,251,2,0,0,14,99,111,108,100,95,109,111,100,101,95,104,97,115,104,32,102,54,48,98,52,102,102,99,101,56,49,49,56,54,53,50,101,57,98,55,56,48,48,100,52,49,53,53,51,102,56,55,0,15,98,108,117,101,95,115,99,101,110,101,95,104,97,115,104,32,57,52,53,54,57,99,53,50,98,51,100,57,99,51,101,57,100,52,97,99,99,99,99,50,99,99,51,100,99,55,102,50,254,3,251,3,1,0,15,98,108,117,101,95,104,97,108,108,95,99,111,117,110,116,193,235,1,252,58,188,124,214,153,1,0,0,0,15,99,111,108,100,95,98,97,115,101,95,99,111,117,110,116,193,159,3,0,13,111,100,100,95,104,105,108,108,95,104,97,115,104,32,102,53,55,48,54,98,48,53,98,51,99,98,50,98,99,102,99,101,54,52,48,56,101,102,101,50,100,56,48,54,50,51,254,4,251,3,0,0,4,240,159,152,139,64,128,100,57,97,51,51,99,101,52,98,97,98,57,97,50,56,49,55,98,55,98,48,49,57,101,52,50,55,49,97,101,100,102,50,98,49,48,56,54,98,51,52,52,54,52,97,101,100,52,98,51,57,49,56,102,49,98,56,56,56,56,56,49,56,100,100,98,99,50,52,102,51,99,99,54,55,97,99,99,56,55,54,102,52,57,101,56,99,101,48,55,52,53,49,98,50,98,55,98,48,54,50,52,53,97,56,97,51,50,50,102,100,101,100,99,102,53,51,52,57,50,56,50,53,100,56,51,101,102,13,11,99,97,108,109,95,115,117,110,115,101,116,29,29,0,0,0,24,0,0,0,4,0,0,3,50,94,56,5,192,0,1,4,3,50,94,55,5,192,128,0,255,0,16,112,117,114,112,108,101,95,119,111,111,100,95,104,97,115,104,32,55,97,50,52,51,50,99,102,48,99,52,55,53,101,57,99,48,97,50,53,101,49,52,101,48,48,54,50,55,49,100,97,254,11,251,1,0,0,18,115,112,114,105,110,103,95,117,110,105,111,110,95,99,111,117,110,116,193,219,2,254,14,251,1,0,0,16,115,110,111,119,121,95,98,111,97,116,95,99,111,117,110,116,193,251,0,255,58,213,6,148,237,168,189,201])

print("=" * 110)
print("Redis RDB 文件 - 完整Key/Value解析")
print("=" * 110)

def extract_string_at(data, pos):
    """从指定位置提取字符串"""
    if pos >= len(data):
        return None, pos
    
    first_byte = data[pos]
    
    # 长度编码
    if first_byte < 64:  # 6位长度
        length = first_byte
        pos += 1
    elif first_byte < 128:  # 14位长度
        if pos + 1 >= len(data):
            return None, pos
        second_byte = data[pos + 1]
        length = ((first_byte & 0x3F) << 8) | second_byte
        pos += 2
    elif first_byte == 0x80:  # 32位长度
        if pos + 4 >= len(data):
            return None, pos
        length = struct.unpack('>I', data[pos+1:pos+5])[0]
        pos += 5
    else:
        return None, pos
    
    if pos + length > len(data):
        return None, pos
    
    string = data[pos:pos+length].decode('latin1', errors='ignore')
    return string, pos + length

# 第一阶段：找所有标记
print("\n[第一阶段] 扫描文件结构\n")

pos = 0
# 跳过文件头
header = data[0:9]
print(f"文件头 (位置 0-8): {header.decode('ascii')}\n")
pos = 9

db_positions = []
metadata = []
record_count = 0

while pos < len(data):
    byte = data[pos]
    
    if byte == 0xFA:  # 辅助字段
        print(f"位置 {pos:4d}: 0xFA 辅助字段标记")
        pos += 1
        key, pos = extract_string_at(data, pos)
        value, pos = extract_string_at(data, pos)
        print(f"           {key} = {value}")
        metadata.append((key, value))
    
    elif byte == 0xFE:  # 数据库选择
        db_id = data[pos + 1]
        print(f"\n位置 {pos:4d}: 0xFE 数据库选择器 -> DB {db_id}")
        db_positions.append((pos, db_id))
        pos += 2
    
    elif byte == 0xFB:  # 哈希表大小
        size_hash = data[pos + 1]
        size_expire = data[pos + 2]
        print(f"位置 {pos:4d}: 0xFB 哈希表信息 (hash_size={size_hash}, expire_size={size_expire})")
        pos += 3
    
    elif byte == 0xFF:  # 文件结束
        print(f"\n位置 {pos:4d}: 0xFF 文件结束标记")
        pos += 1
        checksum = data[pos:pos+8]
        print(f"        校验和: {checksum.hex()}")
        break
    
    elif byte == 0xFD:  # 秒级过期时间
        print(f"位置 {pos:4d}: 0xFD 过期时间标记 (秒)")
        exp_time = struct.unpack('<I', data[pos+1:pos+5])[0]
        print(f"        过期时间: {exp_time}")
        pos += 5
        # 接下来是value type
        if pos < len(data):
            value_type = data[pos]
            print(f"        值类型: 0x{value_type:02x}")
            pos += 1
            record_count += 1
    
    elif byte == 0xFC:  # 毫秒级过期时间
        print(f"位置 {pos:4d}: 0xFC 过期时间标记 (毫秒)")
        exp_time = struct.unpack('<Q', data[pos+1:pos+9])[0]
        print(f"        过期时间: {exp_time}")
        pos += 9
        if pos < len(data):
            value_type = data[pos]
            print(f"        值类型: 0x{value_type:02x}")
            pos += 1
            record_count += 1
    
    elif byte < 32 or (byte >= 0xC0 and byte < 0xFA):
        # 可能是值类型或特殊编码
        print(f"位置 {pos:4d}: 字节 0x{byte:02x} ({byte})")
        pos += 1
    
    else:
        pos += 1

print("\n" + "=" * 110)
print("[第二阶段] 提取所有Key/Value对\n")

# 重新解析，这次专注于Key/Value
pos = 9  # 跳过文件头
current_db = 0
all_data = {}

print(f"{'数据库':<8} {'Key':<40} {'值类型':<10} {'值':<50}")
print("-" * 110)

while pos < len(data):
    byte = data[pos]
    
    if byte == 0xFA:
        pos += 1
        key, pos = extract_string_at(data, pos)
        value, pos = extract_string_at(data, pos)
    
    elif byte == 0xFE:
        current_db = data[pos + 1]
        all_data[current_db] = []
        pos += 2
    
    elif byte == 0xFB:
        pos += 3
    
    elif byte == 0xFF:
        break
    
    elif byte == 0xFD:
        pos += 5
        if pos < len(data):
            value_type = data[pos]
            pos += 1
            # 读取key
            key, pos = extract_string_at(data, pos)
            # 读取value
            value, pos = extract_string_at(data, pos)
            
            if key and value:
                type_name = {0: 'String', 1: 'List', 2: 'Set', 3: 'ZSet', 4: 'Hash'}.get(value_type, f'Type{value_type}')
                print(f"DB {current_db:<6} {key:<40} {type_name:<10} {str(value)[:50]:<50}")
                if current_db not in all_data:
                    all_data[current_db] = []
                all_data[current_db].append({'key': key, 'value': value, 'type': type_name})
    
    elif byte == 0xFC:
        pos += 9
        if pos < len(data):
            value_type = data[pos]
            pos += 1
            key, pos = extract_string_at(data, pos)
            value, pos = extract_string_at(data, pos)
            
            if key and value:
                type_name = {0: 'String', 1: 'List', 2: 'Set', 3: 'ZSet', 4: 'Hash'}.get(value_type, f'Type{value_type}')
                print(f"DB {current_db:<6} {key:<40} {type_name:<10} {str(value)[:50]:<50}")
                if current_db not in all_data:
                    all_data[current_db] = []
                all_data[current_db].append({'key': key, 'value': value, 'type': type_name})
    
    else:
        # 直接key-value对
        key, new_pos = extract_string_at(data, pos)
        if key:
            pos = new_pos
            value, pos = extract_string_at(data, pos)
            if value:
                print(f"DB {current_db:<6} {key:<40} {'String':<10} {str(value)[:50]:<50}")
                if current_db not in all_data:
                    all_data[current_db] = []
                all_data[current_db].append({'key': key, 'value': value, 'type': 'String'})
        else:
            pos += 1

print("\n" + "=" * 110)
print("汇总\n")
for db_id in sorted(all_data.keys()):
    print(f"【数据库 {db_id}】包含 {len(all_data[db_id])} 个Key:")
    for item in all_data[db_id]:
        print(f"  • {item['key']:<40} ({item['type']:>6}) = {item['value'][:60]}")
    print()

print("=" * 110)