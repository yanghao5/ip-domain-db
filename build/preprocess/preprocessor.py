import json
import os

def convert_json(file_path):
    with open(file_path, 'r') as f:
        data = json.load(f)

    for rule in data.get('rules', []):            
        if isinstance(rule.get('domain'), str):
            rule['domain'] = [rule['domain']]  
            
        if isinstance(rule.get('domain_suffix'), str):
            rule['domain_suffix'] = [rule['domain_suffix']]  

        if isinstance(rule.get('domain_keyword'), str):
            rule['domain_keyword'] = [rule['domain_keyword']]
            
        if isinstance(rule.get('domain_regex'), str):
            rule['domain_regex'] = [rule['domain_regex']] 
            
        if isinstance(rule.get('ip_cidr'), str):
            rule['ip_cidr'] = [rule['ip_cidr']] 
        

    with open(file_path, 'w') as f:
        json.dump(data, f, indent=2)

def process_directory(directory_path):
    # 获取目录下所有的 .json 文件
    for root, dirs, files in os.walk(directory_path):
        for file in files:
            if file.endswith('.json'):
                file_path = os.path.join(root, file)
                convert_json(file_path)


if __name__ == '__main__':    
    process_directory("meta-rules-dat/geo-lite/geosite")
    process_directory("meta-rules-dat/geo/geosite")
    process_directory("meta-rules-dat/geo-lite/geoip")
    process_directory("meta-rules-dat/geo/geoip")
    process_directory("meta-rules-dat/asn")
    
