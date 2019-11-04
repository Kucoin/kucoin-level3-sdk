import json
import socket


class RPC(object):
    __slots__ = ('port', 'host', 'token', 'conn')

    def __init__(self, host: str, port: int, token: str):
        """
        :param host:
        :param port:
        :param token:
        """
        self.port = port
        self.host = host
        self.token = token
        """
        create tcp connect
        """
        self.conn = socket.create_connection((self.host, self.port))

    def read(self) -> str:
        """
        :return:
        """
        rev = b''
        while True:
            c = self.conn.recv(1)
            if c == b'\n' or c == b'':
                break
            else:
                rev += c
        return rev.decode("utf-8")

    def execute(self, data: map) -> map:
        """
        :param data:
        :return:
        """
        data['id'] = 0

        self.conn.sendall(json.dumps(data).encode())

        response = json.loads(self.read())

        if response.get('id') != 0:
            raise Exception("expected id=%s, received id=%s: %s" % (0, response.get('id'), response.get('error')))

        if response.get('error') is not None:
            raise Exception(response.get('error'))

        result = json.loads(response.get('result'))

        if result['code'] != '0':
            raise Exception("rpc execute fail: %s" % result['error'])

        return result

    def close(self):
        """
        :return:
        """
        self.conn.close()

    def call(self, method: str, **kwargs):
        """
        :param method:
        :param kwargs:
        :return:
        """
        params = {
            'token': self.token,
        }
        if kwargs:
            params.update(kwargs)

        data = {
            'method': "Server.%s" % method,
            'params': [params],
        }

        return self.execute(data)

    def get_ticker(self, number):
        result = self.call("GetPartOrderBook", number=number)
        ticker = json.loads(result['data'])
        if ticker['sequence'] == 0:
            raise Exception("rpc get ticker fail: sequence is null")
        return ticker

    def get_all_ticker(self):
        result = self.call("GetOrderBook")
        ticker = json.loads(result['data'])
        if ticker['sequence'] == 0:
            raise Exception("rpc get all ticker fail: sequence is null")
        return ticker

    def add_event_order_id(self, data, channel):
        args = {}
        for i in data:
            args[i] = [channel]
        return self.call("AddEventOrderIdsToChannels", data=args)

    def add_event_client_id(self, data, channel):
        args = {}
        for i in data:
            args[i] = [channel]
        return self.call("AddEventClientOidsToChannels", data=args)
