import socket
import logging
import signal
import sys
from common.utils import Contestant, is_winner, recv_all

class BadProtocolError(Exception): ...

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._client_socket = None
        signal.signal(signal.SIGTERM, self.__sigterm_handler)

    def __sigterm_handler(self, received_signal, _):
        if received_signal != signal.SIGTERM: return
        logging.info('Closing accepter socket...')
        self._server_socket.close()
        logging.info('Closing client socket...')
        if self._client_socket != None and self._client_socket.fileno() != -1:
            self._client_socket.close()
        logging.info('Succesfully freed resources')
        sys.exit(0)

    def run(self):
        while True:
            self.__accept_new_connection()
            self.__handle_client_connection()

    def __receive_contestant(self) -> Contestant:
        dataLength = int.from_bytes(recv_all(self._client_socket, 2), byteorder='big', signed=False) # Get total bytes to receive
        data = recv_all(self._client_socket, dataLength).decode('utf-8').split(';')
        if len(data) != 4: raise BadProtocolError()
        return Contestant(*data)

    def __send_contestant_result(self, is_winner_contestant: bool):
        res = b'\x01' if is_winner_contestant else b'\x00'
        self._client_socket.sendall(res)

    def __handle_client_connection(self):
        try:
            contestant = self.__receive_contestant()
            logging.info(
                'Received the following contestant {}'
                .format(vars(contestant)))
            is_winner_contestant = is_winner(contestant)
            self.__send_contestant_result(is_winner_contestant)
        except OSError:
            logging.error("Error while reading socket {}".format(self._client_socket))
        except BadProtocolError:
            logging.error("Error while communicating with client: bad protocol")
        except Exception as err:
            logging.error("Error while communicating with client: {}".format(repr(err)))
        finally:
            self._client_socket.close()

    def __accept_new_connection(self):
        logging.info("Proceed to accept new connections")
        self._client_socket, addr = self._server_socket.accept()
        self._client_socket.settimeout(5.0)
        logging.info('Got connection from {}'.format(addr))
