from dataclasses import dataclass
import random
import sys
import string
import time
import datetime

random.seed(time.time())

@dataclass
class PersonData():
  name: str
  surname: str
  document: str
  birthdate: str

  def get_random():
    names = ['Julio', 'Mateo', 'Claudio', 'Pablo', 'Juan', 'Federico']
    surnames = ['Gutierrez', 'Rodriguez', 'Alvarez', 'Bonifetto', 'Ortiz', 'Gallo']
    return PersonData(
      name=random.choice(names),
      surname=random.choice(surnames),
      document=''.join(random.choice(string.digits) for _ in range(8)),
      birthdate=datetime.date.fromtimestamp(random.randint(1, int(time.time()))).strftime('%Y-%m-%d')
    )

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
    volumes:
      - ./server/config.ini:/config.ini
"""
)

for i in range(1, clientCount + 1):
  pd = PersonData.get_random()
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
      - CLI_PERSON_NAME={name}
      - CLI_PERSON_SURNAME={surname}
      - CLI_PERSON_DOCUMENT={document}
      - CLI_PERSON_BIRTHDATE={birthdate}
    networks:
      - testing_net
    depends_on:
      - server
    volumes:
      - ./client/config.yaml:/config.yaml
  """.format(clientId=i, name=pd.name, surname=pd.surname, 
            document=pd.document, birthdate=pd.birthdate)
  )

file.write(
  """
  server-health-check:
    container_name: server-health-check
    image: server-health-check
    entrypoint: /health-check.sh
    build: ./health
    networks:
      - testing_net
    depends_on:
      - server
  """
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

