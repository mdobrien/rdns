import argparse
import ipaddress
import json
import logging
import math
import os
import sys
import time
# from scapy.all import *

# ----------------------------------------
import logging
logging.basicConfig(
    level=logging.DEBUG,
    format="%(asctime)s [%(levelname)s] %(message)s",
    handlers=[
        logging.FileHandler("debug.log"),
        logging.StreamHandler()
    ]
)
# ----------------------------------------
# GLOBALS
DATA_DIR = '/data/'
"""


# Sucess 1
# pre dynamicly adjust wait based on interval time
python3 rdns.py run --cidr 128.8.0.0/16 --resolvers 8.8.8.8 --qps 500
executed 65500, last ip: 128.8.255.219, results size: 0.58992mb, interval: 48.662986278533936s
avg name length: 21.842153441272384
#names/#queries: 17479 / 65536
#timeouts: 5
#noname: 48052
Elapse time sending DNS queries: 12325.574210643768s

# Success 2
python3 rdns.py run --cidr 128.8.0.0/16 --resolvers 8.8.8.8
avg name length: 21.846541745458712
#names/#queries: 17451 / 65536
#timeouts: 100
#noname: 47985
Elapse time sending DNS queries: 4618.8515367507935s
writing results to /data/rdns.json
Elapse time writing results to disk: 0.05539727210998535s
wrote files to disk
runtime: 4619.0649337768555

#Test 3 - aiodns
	runs at ~33 qps
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
    yes    /16   65,536
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
# Todo: determine max QPS with no rate limiting for current config


"""
Situation:
 cloudflare will sometimes respond No such name and say this is the SOA for that IP. 
 Why does this happen? When does this happen? Is this cloudflare rate limiting or is this because the ip had no name
 I could add logic to query the SOA directly instead of cloudflare.

"""


def cidr_to_24(cidr):
    # Parse the input CIDR string
    network = ipaddress.ip_network(cidr)
    
    # Create a list of /24 CIDRs contained within the input CIDR
    cidr_list = []
    for subnet in network.subnets(new_prefix=24):
        cidr_list.append(str(subnet)[:-4])
    
    return cidr_list

def expand_cidr(cidr):
    """
    Expands a CIDR notation for an IPv4 network to a list of IP addresses
    Example: "192.168.0.0/28" -> ['192.168.0.0', '192.168.0.1', '192.168.0.2', ...]
    """
    network = ipaddress.IPv4Network(cidr)
    addresses = [str(ip) for ip in network]
    return addresses


def generate_run_id():
	"""

	Returns:
		(str) : First 10 chars of hex of uuid4 string
	"""
	import datetime
	# print (f'{dir(datetime.datetime.now())}')
	date = str(datetime.datetime.now())
	date = date.replace(' ', '-')
	idx = date.find('.')
	date = date[5:idx]
	logging.info(f'{date}')
	return date
	# return uuid.uuid4().hex[:10]





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
	wait = None
	if qps:
		wait = 60 / qps
	# breakpoint()
	# log.debug(f'Sleep interval: {wait!r}s')
	logging.debug(f'Sleep interval: {wait}s')


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
			# logging.debug (f'name: {type(name)}, {name}')
			# logging.debug (f'net: {type(net)}, {net}')
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
			# Todo add qps/qpm metrics and #nonames, #timeouts, #names
			logging.debug (f'executed {executed_queries}, last ip: {net}, results size: {sys.getsizeof(results) / 1000000}mb, interval: {curr_time-prev_time}s')
			interval_time = curr_time - prev_time

			# update wait
			if qps:
				real_spq = interval_time / interval_queries
				desired_spq = 60 / qps
				if real_spq > desired_spq:				
					delta_spq = real_spq - desired_spq
					wait = wait - (delta_spq)
					logging.debug (f'Sleep interval: {wait}s')
				else:
					delta_spq = desired_spq - real_spq 
					wait = wait + (delta_spq)
					logging.debug (f'Sleep interval: {wait}s')


			prev_time = time.time()

		if qps:
			time.sleep(abs(wait))

	# get id used for run
	run_id = generate_run_id()

	# logging.debug (f'{results}')
	if len(results) > 0:
		# TODO: add qps and qpm metric
		logging.debug (f'avg name length: {total_name_chars / len(results)}')
		logging.debug (f'#names/#queries: {len(results)} / {len(nets)}')
		logging.debug (f'#timeouts: {len(failed_rdns_ips)}')
		logging.debug (f'#noname: {len(noname_rdsn_ips)}')
	else:
		logging.debug (f'#results: 0')
	logging.debug(f'Elapse time sending DNS queries: {time.time() - st:.2f}s')

	store_dns(failed_rdns_ips=failed_rdns_ips, noname_rdsn_ips=noname_rdsn_ips, results=results, run_id=run_id)

def store_dns(failed_rdns_ips=None, noname_rdsn_ips=None, results=None, run_id=None):
	"""
	Convert result data to json and write to disk. Take the results of the all the reverse dns queries and write them to disk.
	A directory is created in /data/ that will contain 3 files. 

	/data/run_id/rdns.json -> is a JSON dumo of the results dict AKA rdns resutls of CIDRS looked up
	/data/run_id/failed.txt -> list ip IP address where the DNS query timed out w/o a response
	/data/run_id/noname.txt -> List of IPs that have no name or a .arpa name	
	
	Args:
		results (dict) : Keys is ip address and value is hostname of IP
	
	Returns:
		bool: hard coded True value
	"""
	path = DATA_DIR + run_id + '/'

	logging.debug (f'writing results to {path}')
	file = 'rdns.json'
	st = time.time()	
	results_json = json.dumps(results)

	if not os.path.exists(path):
		os.mkdir(path)


	with open(path + file, 'w') as f:
		f.write(results_json)

	logging.debug (f'Elapse time writing results to disk: {time.time() - st:.2f}s')

	failed_rdns_ips_path = path + 'failed.txt'
	noname_rdns_ips_path = path + 'noname.txt'


	with open(failed_rdns_ips_path, 'w') as f:
		f.write(str(failed_rdns_ips))

	with open(noname_rdns_ips_path, 'w') as f:
		f.write(str(noname_rdsn_ips))

	return True

def generate_tasking(cidrs, resolvers):
	# split cidrs to /24s
	# split nets equally among resolvers
	nets = []
	for cidr in cidrs:
		nets += cidr_to_24(cidr)
	# logging.debug(f'size={len(nets)} nets={nets}')

	pivot = math.ceil(len(nets) / len(resolvers))

	tasking = dict()
	start = 0
	end = pivot
	for resolver in resolvers:
		partition = nets[start:end]
		# logging.debug(f'start={start}, end={end} size={len(partition)} partition={partition}')

		start += pivot
		end += pivot
		if partition:
			tasking[resolver] = partition
		else:
			logging.info(f'No tasking sent to {resolver}')

		# logging.debug(f'{tasking!r}')
	tasking = json.dumps(tasking)

	return tasking


def godns(cidrs, resolvers, workers, qps):

	# TODO: handle more than one CIDR
	# TODO: create work message
	# TODO: sent tasking message to worker from master
	# Todo: recv tasking msg and parse it
	# Todo: dispacht tasking to multhreading asynchronous rdns lookup
	if len(cidrs) > 0 and len(resolvers):

		tasking = generate_tasking(cidrs, resolvers)

		with open('/root/tasking.json', 'w') as f:
			f.write(tasking)

		# logging.info(f'{tasking!r}')
		# cidr = cidrs[0]
		cmd = f'go run /root/godns/dns.go'
		os.system(cmd)
		logging.info(f'Execute: {cmd}')


# ------------------------------------------------------------------------
def cli_rdns_go(args):
	cidrs = args.cidrs
	resolvers = args.resolvers
	qps = args.qps
	workers = args.workers

	start = time.time()
	godns(cidrs, resolvers, workers, qps)
	logging.debug (f'runtime: {time.time() - start:.2f}')
# ------------------------------------------------------------------------
# def cli_rdns(args):
# 	"""
# 	Invoed by running rdns from the commandline
# 	"""
# 	cidrs = args.cidrs
# 	resolvers = args.resolvers
# 	qps = args.qps
# 	workers = args.workers

# 	start = time.time()
# 	# conf.verb = 0 # disable scapy debug statements
# 	logging.debug(f'cidr={args.cidrs}, dns_server_ip={args.resolvers}, qps={args.qps}, workers={args.workers}')
# 	# Todo  master - add num clients as param that gets parsed
# 	# TODO: master - generate tasking messages
# 	# TODO: master - init worker nodes
# 	# TODO: master - send tasking msg to worker nodes
# 	# TODO: master - establisg cnx to worker nodes
# 	# TODO: master - send tasking to worker nodes
# 	# rdns(cidr=args.cidr, dns_server_ip=args.resolvers, qps=args.qps)
# 	generate_tasking(cidrs, resolvers, workers, qps)

# 	logging.debug (f'runtime: {time.time() - start:.2f}')

def cli_benchmark(args):
	filepath = args.file
	# with open(filepath, "r") as f:
	# ips = []
	# domains = []
	# resolvers = parse resolvers
	# for resolver in resolvers
		 # look up ips
		 # look up domains
		 # save latency for each query
		 # calc number of correct answers
	# print stats


def cli_clean(args):
	"""
		This function will remove all directories in DATA_HOME
		except for files specified to be keep 
		example invocation: python3 rdns.py clean keep dir1 dir 2 
		This will remove all dirs in DATA_HOME except dir1 and dir2
	Args:
	Returns:
		Nothing
	"""
	logging.info(f'data_dir: {DATA_DIR}')
	if os.path.exists(DATA_DIR):
		contents = os.listdir(DATA_DIR)
		for folder in contents:
			path = DATA_DIR + folder + '/'
			logging.info(f'data_dir: {path}')
			# breakpoint()
			if folder in args.keep:
				for file in os.listdir(path):
					os.remove(path + file)
					logging.debug(f'removed: {path+file}')
				os.rmdir(path)
				logging.debug(f'removed: {path}')

def cli_test(args):
	if os.path.exists(DATA_DIR):
			contents = os.listdir(DATA_DIR)
			for folder in contents:
				path = DATA_DIR + folder + '/'
				logging.info(f'data_dir: {path}')
				for file in os.listdir(path):
					logging.debug(f'{path + file}')
					os.remove(path + file)
				# os.rmdir(path)
				breakpoint()
	
	pass

# ------------------------------------------------------------------------

if __name__ == "__main__":

	    
    parser = argparse.ArgumentParser()
    subparser = parser.add_subparsers()

    rdns_parser = subparser.add_parser('benchmark')
    rdns_parser.add_argument('-f', '--file', type=str)
    rdns_parser.set_defaults(func=cli_benchmark)

    rdns_parser = subparser.add_parser('clean')
    rdns_parser.add_argument('--keep', nargs='+', required=False)
    rdns_parser.set_defaults(func=cli_clean)

    rdns_parser = subparser.add_parser('run')
    rdns_parser.add_argument('-d','--resolvers', nargs='+', required=False)
    rdns_parser.add_argument('-c','--cidrs', nargs='+', required=False)
    rdns_parser.add_argument('-q','--qps', type=int)
    rdns_parser.add_argument('-w','--workers', type=int)
    rdns_parser.set_defaults(func=cli_rdns_go)


    rdns_parser = subparser.add_parser('test')
    rdns_parser.set_defaults(func=cli_test)
    
    args = parser.parse_args()
    # -----------------------------------
    if hasattr(args, 'func'):
        args.func(args)
    else:
        parser.logging.debug_help()
    # -----------------------------------
    sys.exit()
#--------------------------------------------------------------------------------

