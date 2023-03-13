import ipaddress
import time
import argparse
import json
import sys
from scapy.all import *

"""
# pre dynamicly adjust wait based on interval time
python3 rdns.py run --cidr 128.8.0.0/16 --destination 8.8.8.8 --qps 500
executed 65500, last ip: 128.8.255.219, results size: 0.58992mb, interval: 48.662986278533936s
avg name length: 21.842153441272384
#names/#queries: 17479 / 65536
#timeouts: 5
#noname: 48052
Elapse time sending DNS queries: 12325.574210643768s
"""
"""
 SUCCEEDED CIDRS #ips
	yes    /32   1
    yes    /31   2
    yes    /30   4
    yes    /29   8
    yes    /28   16
    yes    /27   32
    yes    /26   64
    yes    /25   128
    yes    /24   256
    yes    /23   512
    yes    /22   1,024
    yes    /21   2,048
    yes    /20   4,096
    yes    /19   8,192
    yes    /18   16,384
    yes    /17   32,768
    N/A    /16   65,536
    N/A    /15   131,072
    N/A    /14   262,144
    N/A    /13   524,288
    N/A    /12   1,048,576
    N/A    /11   2,097,152
    N/A    /10   4,194,304
    N/A    /9    8,388,608
    N/A    /8    16,777,216
    N/A    /7    33,554,432
    N/A    /6    67,108,864
    N/A    /5    134,217,728
    N/A    /4    268,435,456
    N/A    /3    536,870,912
    N/A    /2    1,073,741,824
    N/A    /1    2,147,483,648
    N/A    /0    4,294,967,296
"""


# Todo: create a logger that wraps python logging
# Todo: add catch for ctrl-c which will dump results thus far and terminate
# Todo: check if response says no such name are if it tells you the DNS server you should query
# Todo: #names, #timeout, #nonames, #queries  to interval stats
# Todo: dyncamically update wait interval based on interval time to factor in network/compute time in to wait
# Todo: handle case where querying local DNS server and rescursive queries not support but the response tells the name of the DNS server to query example: 
		# blox.net.umd.edu is a main umd dns server with ip 128.8.76.12
		# dig @128.8.76.12 -x 128.8.73.251 will give response saying the names of the SOAs for 128.8.73.251 because recursive queries are not supported

"""
Situation:
 cloudflare will sometimes respond No such name and say this is the SOA for that IP. 
 Why does this happen? When does this happen? Is this cloudflare rate limiting or is this because the ip had no name
 I could add logic to query the SOA directly instead of cloudflare.

"""



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
		print (f'Sleep interval: {wait}s')

	total_name_chars = 0
	executed_queries = 0

	nets = expand_cidr(cidr)
	results = dict()
	st = time.time()
	failed_rdns_ips = list()
	noname_rdsn_ips = list()
	prev_time = time.time()
	for net in nets:
		# TODO: dynamically throttle request rate by qps parameter
		qname = ipaddress.ip_address(net).reverse_pointer  # convert ip address to reverse pointer
		packet = IP(dst=dns_server_ip)/UDP()/DNS(rd=1,qd=DNSQR(qname=qname, qtype='PTR'))
		timeout = 3
		ans = sr1(x=packet, timeout=timeout)

		if not ans:
			# executes if send/recv times out
			failed_rdns_ips.append(net)
		elif ans.an:
			# executes if recv a propper response dns record
			name = ans.an.rdata.decode("utf-8") 
			results[net] = name
			total_name_chars += len(name)
			# print (f'name: {type(name)}, {name}')
			# print (f'net: {type(net)}, {net}')
		else:
			# executes when ip does not have a name in proper dns response
			# Todo: verify the statement above is correct
			noname_rdsn_ips.append(net)

		# Todo
		# handle case where response contains SOA and no name for ip
		interval_queries = 250
		executed_queries += 1
		if executed_queries % interval_queries == 0:
			curr_time = time.time()
			# Todo add real qps 
			print (f'executed {executed_queries}, last ip: {net}, results size: {sys.getsizeof(results) / 1000000}mb, interval: {curr_time-prev_time}s')
			interval_time = curr_time - prev_time

			# update wait
			real_spq = interval_time / interval_queries
			desired_spq = 60 / qps
			if real_spq > desired_spq:				
				delta_spq = real_spq - desired_spq
				wait = wait - (delta_spq)
				print (f'Sleep interval: {wait}s')
			else:
				delta_spq = desired_spq - real_spq 
				wait = wait + (delta_spq)
				print (f'Sleep interval: {wait}s')





			prev_time = time.time()

		if qps:
			time.sleep(wait)

	# print (f'{results}')
	if len(results) > 0:
		print (f'avg name length: {total_name_chars / len(results)}')
		print (f'#names/#queries: {len(results)} / {len(nets)}')
		print (f'#timeouts: {len(failed_rdns_ips)}')
		print (f'#noname: {len(noname_rdsn_ips)}')
	else:
		print (f'#results: 0')
	print(f'Elapse time sending DNS queries: {time.time() - st}s')

	path = '/data/rdns.json'
	log(failed_rdns_ips, noname_rdsn_ips)

	print (f'writing results to {path}')
	store_dns(results=results, path=path)
	print (f'wrote files to disk')

# Todo: create a new directory for each one. Store date, target cidr, dns resolver, plus results
def log(failed_rdns_ips=None, noname_rdsn_ips=None):

	failed_rdns_ips_path = '/data/failed.txt'
	noname_rdns_ips_path = '/data/noname.txt'

	if not os.path.exists('/data/'):
		os.mkdir('/data')

	if os.path.exists(failed_rdns_ips_path):
		os.remove(failed_rdns_ips_path)

	if os.path.exists(noname_rdns_ips_path):
		os.remove(noname_rdns_ips_path)

	with open(failed_rdns_ips_path, 'w') as f:
		f.write(str(failed_rdns_ips))

	with open(noname_rdns_ips_path, 'w') as f:
		f.write(str(noname_rdsn_ips))


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

	if os.path.exists('/data/rdns.json'):
		os.remove('/data/rdns.json')


	with open('/data/rdns.json', 'w') as f:
		f.write(results_json)

	print (f'Elapse time writing results to disk: {time.time() - st}s')

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

