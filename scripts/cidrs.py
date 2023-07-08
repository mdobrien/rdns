import ipaddress

def generate_slash_24s(cidr_list):
    slash_24_list = []
    for cidr in cidr_list:
        network = ipaddress.IPv4Network(cidr)
        for subnet in network.subnets(new_prefix=24):
            slash_24_list.append(str(subnet))
    return slash_24_list

# Example usage
cidr_list = [
    "10.0.0.0/8",
    "100.64.0.0/10",
    "127.0.0.0/8",
    "169.254.0.0/16",
    "172.16.0.0/12",
    "192.0.0.0/24",
    "192.0.2.0/24",
    "192.88.99.0/24",
    "192.168.0.0/16",
    "198.18.0.0/15",
    "198.51.100.0/24",
    "203.0.113.0/24",
    "224.0.0.0/4",
    "240.0.0.0/4"       
]

cidrMap = dict()
result = generate_slash_24s(cidr_list)

for cidr in result:
    cidrMap[cidr] = True
#print(result)