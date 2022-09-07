import multiprocessing
import socket
import logging
import signal
import sys
from multiprocessing import Process, Lock
from common.utils import Contestant, is_winner, persist_winners, recv_all


class BadProtocolError(Exception): ...

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._client_socket = None
        total_workers = max(multiprocessing.cpu_count() - 1, 1)
        file_lock = Lock()
        self._client_handlers = [
            Process(target=self.__run_worker, args=[i+1, file_lock]) for i in range(total_workers)
        ]
        signal.signal(signal.SIGTERM, self.__sigterm_handler_manager)

    def __sigterm_handler_worker(self, received_signal, _):
        if received_signal != signal.SIGTERM: return
        logging.info(f'[SERVER {self._worker_number}] Closing accepter socket...')
        self._server_socket.close()
        logging.info(f'[SERVER {self._worker_number}] Closing client socket...')
        if self._client_socket != None and self._client_socket.fileno() != -1:
            self._client_socket.close()
        logging.info(f'[SERVER {self._worker_number}] Succesfully freed resources')
        sys.exit(0)

    def __sigterm_handler_manager(self, received_signal, _):
        if received_signal != signal.SIGTERM: return
        for ch in self._client_handlers: ch.terminate()

    def run(self):
        for ch in self._client_handlers: ch.start()
        self._server_socket.close() # Only the workers will use it
        for ch in self._client_handlers: ch.join()

    def __run_worker(self, worker_number: int, file_lock: Lock):
        self._worker_number = worker_number
        self._file_lock = file_lock
        signal.signal(signal.SIGTERM, self.__sigterm_handler_worker)
        while True:
            self.__accept_new_connection()
            self.__handle_client_connection()

    def __receive_contestants(self) -> list[Contestant]:
        totalContestants = int.from_bytes(recv_all(self._client_socket, 2), byteorder='big', signed=False)
        if totalContestants == 0: return []
        contestants = []
        for _ in range(totalContestants):
            dataLength = int.from_bytes(recv_all(self._client_socket, 2), byteorder='big', signed=False)
            data = recv_all(self._client_socket, dataLength).decode('utf-8').split(';')
            if len(data) != 4: raise BadProtocolError()
            contestants.append(Contestant(*data))
        return contestants

    def __send_winners(self, winners: list[Contestant]):
        self._client_socket.sendall(len(winners).to_bytes(2, byteorder='big', signed=False))
        for w in winners:
            winnerData = bytearray(';'.join([
                w.first_name,
                w.last_name,
                w.document,
                w.birthdate.strftime('%Y-%m-%d')]
                ).encode('utf-8'))
            self._client_socket.sendall(len(winnerData).to_bytes(2, byteorder='big', signed=False))
            self._client_socket.sendall(winnerData)

    def __handle_client_connection(self):
        try:
            contestants = self.__receive_contestants()
            while contestants:
                logging.info(f'[SERVER {self._worker_number}] Received contestants batch')
                winners = list(filter(lambda c: is_winner(c), contestants))
                self.__log_winners(winners)
                self.__send_winners(winners)
                contestants = self.__receive_contestants()
        except OSError:
            logging.error("Error while reading socket {}".format(self._client_socket))
        except BadProtocolError:
            logging.error("Error while communicating with client: bad protocol")
        except Exception as err:
            logging.error("Error while communicating with client: {}".format(repr(err)))
        finally:
            self._client_socket.close()

    def __accept_new_connection(self):
        logging.info(f'[SERVER {self._worker_number}] Proceed to accept new connections')
        self._client_socket, addr = self._server_socket.accept()
        self._client_socket.settimeout(5.0)
        logging.info(f'[SERVER {self._worker_number}] Got connection from {addr}')

    def __log_winners(self, winners: list[Contestant]):
        with self._file_lock:
            persist_winners(winners)
