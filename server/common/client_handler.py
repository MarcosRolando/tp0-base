from dataclasses import dataclass
from multiprocessing import Lock, Array, Value
import signal
import logging
import sys
import socket
from enum import Enum
from common.utils import Contestant, is_winner, persist_winners, recv_all

class RequestType(Enum):
    BATCH = 1
    RESULTS = 2


@dataclass
class ClientHandlerArgs:
    server_socket: socket.socket
    file_lock: Lock
    result_flags: Array # Boolean array
    total_winners: Value # Int

class BadProtocolError(Exception): ...


class ClientHandler:
    def __init__(self, worker_number: int, args: ClientHandlerArgs):
        self._client_socket = None
        self._server_socket = args.server_socket
        self._file_lock = args.file_lock
        self._result_flags = args.result_flags
        self._total_winners = args.total_winners
        self._worker_number = worker_number
        signal.signal(signal.SIGTERM, self.__sigterm_handler)

    def __sigterm_handler(self, received_signal, _):
        if received_signal != signal.SIGTERM: return
        logging.debug(f'[SERVER {self._worker_number} Received SIGTERM signal, terminating server...')
        logging.debug(f'[SERVER {self._worker_number}] Closing accepter socket...')
        self._server_socket.close()
        logging.debug(f'[SERVER {self._worker_number}] Closing client socket...')
        if self._client_socket != None and self._client_socket.fileno() != -1:
            self._client_socket.close()
        logging.debug(f'[SERVER {self._worker_number}] Succesfully freed resources')
        sys.exit(0)

    # Actual entrypoint
    def run(worker_number: int, args: ClientHandlerArgs):
        client_handler = ClientHandler(worker_number, args)
        client_handler._run()

    def _run(self):
        while True:
            self.__accept_new_connection()
            self.__handle_client_connection()


    def __receive_contestants(self) -> list[Contestant]:
        totalContestants = int.from_bytes(recv_all(self._client_socket, 2), byteorder='big', signed=False)
        if totalContestants == 0: raise ValueError()
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

    def __handle_winners_result(self):
        waiting_count = 0 # Init
        with self._result_flags.get_lock(): 
            waiting_count = len(list(filter(lambda f: f == 1, self._result_flags)))
            waiting_for_others = waiting_count > 0
        if waiting_for_others:
            self._client_socket.sendall(b'\x00') # Notify that it will also receive amount of agencies still processing
            self._client_socket.sendall(waiting_count.to_bytes(2, byteorder='big', signed=False))
            self._client_socket.sendall(self._total_winners.value.to_bytes(4, byteorder='big', signed=False))
        else:
            self._client_socket.sendall(b'\x01') # Notify that it will only receive the total amount of winners
            self._client_socket.sendall(self._total_winners.value.to_bytes(4, byteorder='big', signed=False))

    def __handle_batch(self):
        contestants = self.__receive_contestants()
        logging.info(f'[SERVER {self._worker_number}] Received contestants batch')
        winners = list(filter(lambda c: is_winner(c), contestants))
        self.__log_winners(winners)
        winners_count = len(winners)
        self.__send_winners(winners)
        with self._total_winners.get_lock(): self._total_winners.value += winners_count
        with self._result_flags.get_lock(): self._result_flags[self._worker_number-1] = 0 # Reset flag

    def __handle_client_connection(self):
        try:
            reqType = self.__get_request_type()
            if reqType == RequestType.BATCH:    
                with self._result_flags.get_lock(): self._result_flags[self._worker_number-1] = 1 # Processing batch
                self.__handle_batch()
            else:
                self.__handle_winners_result()

        except OSError:
            logging.error("Error while reading socket {}".format(self._client_socket))
        except BadProtocolError:
            logging.error("Error while communicating with client: bad protocol")
        except Exception as err:
            logging.error("Error while communicating with client: {}".format(repr(err)))
        finally:
            self._client_socket.close()
            with self._result_flags.get_lock(): self._result_flags[self._worker_number-1] = 0 # Reset flag


    def __get_request_type(self) -> RequestType:
        reqType = int.from_bytes(recv_all(self._client_socket, 1), byteorder='big', signed=False)
        if reqType != 0 and reqType != 1: raise BadProtocolError()
        return RequestType.BATCH if reqType == 0 else RequestType.RESULTS

    def __accept_new_connection(self):
        logging.info(f'[SERVER {self._worker_number}] Proceed to accept new connections')
        self._client_socket, addr = self._server_socket.accept()
        self._client_socket.settimeout(30.0)
        logging.info(f'[SERVER {self._worker_number}] Got connection from {addr}')

    def __log_winners(self, winners: list[Contestant]):
        with self._file_lock:
            persist_winners(winners)
