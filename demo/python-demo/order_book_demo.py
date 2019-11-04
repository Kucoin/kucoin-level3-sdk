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

    while True:
        rpc = RPC(rpc_config['host'], rpc_config['port'], rpc_config['token'])

        data = rpc.get_ticker(100)

        asks = data['asks']
        bids = data['bids']

        price_list = [{}, {}]

        for ask in asks:
            if ask[1] not in price_list[0].keys():
                price_list[0].update({ask[1]: Decimal(ask[2])})
            else:
                price_list[0].update({
                    ask[1]: price_list[0][ask[1]] + Decimal(ask[2])
                })
            if len(price_list[0]) >= 13:
                price_list[0].pop(ask[1])
                break

        for bid in bids:
            if bid[1] not in price_list[1].keys():
                price_list[1].update({bid[1]: Decimal(bid[2])})
            else:
                price_list[1].update({
                    bid[1]: price_list[1][bid[1]] + Decimal(bid[2])
                })
            if len(price_list[1]) >= 13:
                price_list[1].pop(bid[1])
                break

        d1 = sorted(price_list[0].items(), key=lambda v: v[0], reverse=True)
        d2 = sorted(price_list[1].items(), key=lambda v: v[0], reverse=True)

        os.system(cmd)

        for d in d1:
            print("{} => {}".format(d[0], d[1]))
        print("---Spread---")
        for d in d2:
            print("{} => {}".format(d[0], d[1]))
        time.sleep(0.5)
