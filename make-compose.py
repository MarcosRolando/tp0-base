import sys

try:
  clientCount = int(sys.argv[1])
  if clientCount <= 0: raise ValueError
except:
  print('Usage: python make-compose.py CLIENT_COUNT\nCreate Compose with CLIENT_COUNT clients (must be greater than 0)')
  sys.exit(1)

file = open('docker-compose-dev.yaml', 'w+')

file.write(
"""version: '3'
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - SERVER_PORT=12345
      - SERVER_LISTEN_BACKLOG=7
      - LOGGING_LEVEL=DEBUG
    networks:
      - testing_net
"""
)

for i in range(1, clientCount + 1):
  file.write(
  """
  client{clientId}:
    container_name: client{clientId}
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID={clientId}
      - CLI_SERVER_ADDRESS=server:12345
      - CLI_LOOP_LAPSE=1m2s
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server
  """.format(clientId=i)
  )

file.write(
"""
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24

"""
)

file.close()

