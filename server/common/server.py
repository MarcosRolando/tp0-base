import socket
import logging
import signal
import sys


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._client_socket = None
        signal.signal(signal.SIGTERM, self.close)

    def close(self, received_signal, _):
        if received_signal != signal.SIGTERM: return
        logging.info('Closing accepter socket...')
        self._server_socket.close()
        logging.info('Closing client socket...')
        if self._client_socket != None and self._client_socket.fileno() != -1:
            self._client_socket.close()
        logging.info('Succesfully freed resources')
        sys.exit(0)

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while True:
            self.__accept_new_connection()
            self.__handle_client_connection()

    def __handle_client_connection(self):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            msg = self._client_socket.recv(1024).rstrip().decode('utf-8')
            logging.info(
                'Message received from connection {}. Msg: {}'
                .format(self._client_socket.getpeername(), msg))
            self._client_socket.send("Your Message has been received: {}\n".format(msg).encode('utf-8'))
        except OSError:
            logging.info("Error while reading socket {}".format(self._client_socket))
        finally:
            self._client_socket.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info("Proceed to accept new connections")
        self._client_socket, addr = self._server_socket.accept()
        logging.info('Got connection from {}'.format(addr))
