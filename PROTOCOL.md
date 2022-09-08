Para este ejercicio, el servidor se encuentra esperando conexiones de clientes que enviaran los datos del participante para luego contestarles con el resultado de la Loteria (si gano o perdio).

Todos los datos numericos enviados en el protocolo son en Big Endian. Cualquier error en el protocolo corta la comunicacion del servidor con el cliente y no hay reintentos.

El protocolo utilizado en este caso es el siguiente:

* Cuando un cliente se conecta al servidor envia en primer lugar 2 bytes indicado la longitud de los datos del participante en bytes. Los datos son concatenados con el caracter separador `;` que utilizara el servidor para poder separar los mismos. Logicamente la longitud enviada considera estos separadores.
* El servidor sabe que al aceptar una conexion debe entonces recibir los 2 bytes que indican la longitud de datos a leer. Una vez recibidos los datos verifica que estos sean validos (es decir, debe haber 4 datos separados por el delimitador `;` y deben poder parsearse correctamente). Caso contrario es un fallo de protocolo y se cierra la conexion.
* Una vez obtenido el resultado del participante el servidor envia un unico byte de respuesta al cliente: 0 en caso de ser un perdedor y 1 en caso de ganador.

Una vez finalizado el ultimo paso el servidor pasa a esperar por la conexion del proximo cliente.
