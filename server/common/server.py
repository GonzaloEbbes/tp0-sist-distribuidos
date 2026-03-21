import socket
import logging


class Server:
    def __init__(self, port, listen_backlog):
        self._shutdown_requested = False
        self._client_socket = None

        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)

    def stop(self):
        self._shutdown_requested = True

        if self._client_socket is not None:
            self._client_socket.close()
            logging.info('action: close_client_socket | result: success')
            self._client_socket = None

        self._server_socket.close()
        logging.info('action: close_server_socket | result: success')

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while not self._shutdown_requested:
            client_sock = self.__accept_new_connection()
            if client_sock is None:
                continue
            self.__handle_client_connection(client_sock)

    def __send_message(self, client_sock, msg_bytes):
        total_sent = 0
        while total_sent < len(msg_bytes):
            sent = client_sock.send(msg_bytes[total_sent:])
            if sent == 0:
                raise OSError("Socket closed before sending full message")
            total_sent += sent

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        self._client_socket = client_sock
        try:
            # TODO: Avoid short-read by receiving until a full message boundary is detected.
            msg_bytes = client_sock.recv(1024)
            if len(msg_bytes) == 0:
                raise OSError("Socket closed before receiving message")

            msg = msg_bytes.rstrip().decode('utf-8')
            addr = client_sock.getpeername()
            logging.info(f'action: receive_message | result: success | ip: {addr[0]} | msg: {msg}')
            self.__send_message(client_sock, "{}\n".format(msg).encode('utf-8'))
        except OSError as e:
            if not self._shutdown_requested:
                logging.error(f"action: receive_message | result: fail | error: {e}")
        finally:
            if self._client_socket is not None:
                self._client_socket.close()
                logging.info('action: close_client_socket | result: success')
                self._client_socket = None

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        try:
            c, addr = self._server_socket.accept()
        except OSError:
            if self._shutdown_requested:
                return None
            raise
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
