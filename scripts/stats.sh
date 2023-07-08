#!/bin/sh

#stats util
# - time start
# - time delta
# - tasks complete
# - total tasks
# - perce of tasks completed


TC=$(grep "Finished" /root/debug.log  | cat -n | tail -n 1 | awk '{print $1}')
TT=$(grep "tasking:" /root/debug.log | awk '{print $4}')
START=$(grep "tasking:" /root/debug.log | awk '{print $4}')                                                #TODO: MODIFY TO UNIX TIME STAMP
CURR=$(grep "Finished" /root/debug.log  | cat -n | tail -n 1 | awk '{print $21}') #TODO: MODIFY TO UNIX TIME STAMP
ST=$(grep "tasking:" /root/debug.log | awk '{print $5}')
END=$(grep "Finished:" /root/debug.log | tail -n 1 | awk '{print $21}')
QPS=$(python3 -c  "print(f'{($TC*256)/(($END - $ST)/1000):0.02f}')")


echo "-------------------------------------------------------------"
echo "Timestamp: $END"
python3 -c "print(f'Computed QPS: {($TC*256)/(($END - $ST)/1000):0.02f}')"
python3 -c "print(f'Elapse time: {(($END-$ST) / 1000) / 60:0.02f} minutes AKA {(($END-$ST) / 1000) /3600:0.02f} hours')"

echo "Start time: $START"
echo "Tasks completed: $TC"
python3 -c "print(f'Queries executed: {($TC*256)}')"
echo "Total tasks: $TT"

python3 -c "print(f'Percent completed: {($TC/$TT)*100:0.02f}%')"
python3 -c "print(f'Tasks remaining: {($TT-$TC)}')"
python3 -c "print(f'Estimated time remainined (assuming $QPS qps) {(( ($TT-$TC) * 256 ) / $QPS)/3600:0.02f} hours remaining')"
echo "-------------------------------------------------------------"

# Todo added compute QPS
# TODO add estimed time to completion assuming sustained QPS