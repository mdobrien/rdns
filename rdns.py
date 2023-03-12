import ipaddress
import time
import argparse
import json
from scapy.all import *


# Todo: create a logger that wraps python logging


def expand_cidr(cidr):
    """
    Expands a CIDR notation for an IPv4 network to a list of IP addresses
    Example: "192.168.0.0/28" -> ['192.168.0.0', '192.168.0.1', '192.168.0.2', ...]
    """
    network = ipaddress.IPv4Network(cidr)
    addresses = [str(ip) for ip in network]
    return addresses


def rdns(cidr='128.8.0.0/24', dns_server_ip='8.8.8.8', qps=None):
	"""
	Executes a reverse dns look up on input cidr ips using specified dns server
	The reverse DNS queries can be rate limited using the qps param

	Args:
		cidr (str): IPV4 network cidr
		dns_server_ip (str): IP of dns server dns query will be executed against
		qps (int) : Desired number of DNS queryies executed per minute

	Returns:
		dict: Composed of key/value pairs s.t. Key=IPV4 address and value is the host name of the ipv4 address
	"""

	# Todo: factor in dns lookup time in to wait time
	# TODO: calculate actually rate dns packets are being sent/recv
	# Assums dns lookup is constant time 
	# wait 60s / qps to get time to pause in between rdns request
	if qps:
		wait = 60 / qps

	total_name_chars = 0

	nets = expand_cidr(cidr)
	results = dict()
	st = time.time()
	for net in nets:
		# TODO: dynamically throttle request rate by qps parameter
		qname = ipaddress.ip_address(net).reverse_pointer  # convert ip address to reverse pointer
		ans = sr1(IP(dst=dns_server_ip)/UDP()/DNS(rd=1,qd=DNSQR(qname=qname, qtype='PTR')))

		if ans.an:
			name = ans.an.rdata.decode("utf-8") 
			results[net] = name

			total_name_chars += len(name)
			# print (f'name: {type(name)}, {name}')
			# print (f'net: {type(net)}, {net}')
		else:
			name = None

		if qps:
			time.sleep(wait)


	# print (f'{results}')
	print (f'avg name length: {total_name_chars / len(results)}')
	print (f'#names/#queries: {len(results)} / {len(nets)}')
	print(f'sent DNS querys for {time.time() - st}s')

	path = '/data/rdns.json'
	print (f'writing results to {path}')
	store_dns(results=results, path=path)
	print (f'wrote files to disk')

def store_dns(results, path='/data/rdns.json'):
	"""
	Convert result data to json and write to disk

	Args:
	    results (dict): Key=IPV4 address and value is the host name of the ipv4 address

	Returns:
		bool: hard coded True value

	"""
	st = time.time()
	results_json = json.dumps(results)

	if not os.path.exists('/data'):
		os.mkdir('/data')

	with open('/data/rdns.json', 'w') as f:
		f.write(results_json)

	print (f' Elapse time writing results to disk: {time.time() - st}s')

	return True



# ------------------------------------------------------------------------
def cli_rdns(args):
	"""
	Invoed by running rdns from the commandline
	"""
	start = time.time()
	conf.verb = 0 # disable scapy debug statements
	rdns(cidr=args.cidr, dns_server_ip=args.destination, qps=args.qps)

	print (f'runtime: {time.time() - start}')

# ------------------------------------------------------------------------

if __name__ == "__main__":

	    
    parser = argparse.ArgumentParser()
    subparser = parser.add_subparsers()

    rdns_parser = subparser.add_parser('run')
    rdns_parser.add_argument('-d','--destination')
    rdns_parser.add_argument('-c','--cidr', type=str)
    rdns_parser.add_argument('-q','--qps', type=int)
    rdns_parser.set_defaults(func=cli_rdns)
    
    args = parser.parse_args()
    # -----------------------------------
    if hasattr(args, 'func'):
        args.func(args)
    else:
        parser.print_help()
    # -----------------------------------
    sys.exit()
#--------------------------------------------------------------------------------

