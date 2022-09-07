import multiprocessing
import socket
import signal
from multiprocessing import Process, Lock
from common.client_handler import ClientHandler

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        total_client_handlers = max(multiprocessing.cpu_count() - 1, 1)
        file_lock = Lock()
        self._client_handlers = [
            Process(
                target=ClientHandler.run,
                args=[self._server_socket, file_lock, i+1]
            ) for i in range(total_client_handlers)
        ]
        signal.signal(signal.SIGTERM, self.__sigterm_handler)

    def __sigterm_handler(self, received_signal, _):
        if received_signal != signal.SIGTERM: return
        for ch in self._client_handlers: ch.terminate()

    def run(self):
        for ch in self._client_handlers: ch.start()
        self._server_socket.close() # Only the workers will use it
        for ch in self._client_handlers: ch.join()
