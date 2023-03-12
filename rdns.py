import ipaddress
import time
import argparse
from scapy.all import *



def expand_cidr(cidr):
    """
    Expands a CIDR notation for an IPv4 network to a list of IP addresses
    Example: "192.168.0.0/28" -> ['192.168.0.0', '192.168.0.1', '192.168.0.2', ...]
    """
    network = ipaddress.IPv4Network(cidr)
    addresses = [str(ip) for ip in network]
    return addresses


def rdns(cidr='128.8.0.0/24', dns_server_ip='8.8.8.8', qps=500):
	# TODO add checks to parameter inputs

	nets = expand_cidr(cidr)
	results = []

	for net in nets:
		# TODO: dynamically throttle request rate by qps parameter

		qname = ipaddress.ip_address(net).reverse_pointer  # convert ip address to reverse pointer
		ans = sr1(IP(dst=dns_server_ip)/UDP()/DNS(rd=1,qd=DNSQR(qname=qname, qtype='PTR')))
		# breakpoint()

		if ans.an:
			name = ans.an.rdata
			results.append((net, name))
		else:
			name = None

		# print (f'{ans!r}')
		# print (f'######### {net} - {name} ##########')

		# if net == '128.8.0.2':
			# break

	print (f'{results}')
	print (f'#names/#queries: {len(results)} / {len(nets)}')

def cli_rdns(args):
	"""

	"""

	start = time.time()
	conf.verb = 0 # disable scapy debug statements
	rdns(cidr=args.cidr, dns_server_ip=args.destination, qps=args.qps)

	print (f'runtime: {time.time() - start}')



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

