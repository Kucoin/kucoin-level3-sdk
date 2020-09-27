import time
from level3.rpc import RPC
from config import rpc_config
from decimal import Decimal
import os
import platform

if __name__ == '__main__':
    cmd = ''
    system_os = platform.system()
    if system_os == 'Windows':
        cmd = 'cls'
    elif system_os == 'Darwin' or system_os == 'Linux':
        cmd = 'clear'
    else:
        raise Exception('unsupported system')

    rpc = RPC(rpc_config['host'], rpc_config['port'], rpc_config['token'])

    while True:
        order_book = rpc.get_order_book(11)
#         import sys, json
#         print(json.dumps(order_book))
#         sys.exit(0)

        os.system(cmd)

        asks = order_book['asks']
        asks.reverse()

        for d in asks:
            print("{} => {}".format(d[0], d[1]))
        print("---Spread---")
        for d in order_book['bids']:
            print("{} => {}".format(d[0], d[1]))
        time.sleep(0.5)
