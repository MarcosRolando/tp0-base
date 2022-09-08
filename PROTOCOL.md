Para este ejercicio, el servidor se encuentra esperando conexiones de clientes que enviaran los datos de los participantes de su dataset correspondiente para luego contestarles con los datos de los participantes ganadores. Esto es procesado en forma de batch.

Todos los datos numericos enviados en el protocolo son en Big Endian. Cualquier error en el protocolo corta la comunicacion del servidor con el cliente y no hay reintentos. El servidor procesa los datos de aun cliente a la vez, mas alla de que el envio de los datos y la respuesta sea en batches.

El protocolo utilizado en este caso es el siguiente:

* Cuando un cliente se conecta al servidor comienza a leer los participantes en forma de batch y envia en primer lugar 2 bytes indicado la longitud de los datos de un participante en bytes. Los datos son concatenados con el caracter separador `;` que utilizara el servidor para poder separar los mismos. Logicamente la longitud enviada considera estos separadores. Esta operacion se repite para cada participante del batch siendo procesado.

* El servidor obtiene en base a los datos recibidos los participantes ganadores y reenvia al cliente los datos correspondientes de cada uno. Primero se envian 2 bytes que indican la cantidad total de participantes ganadores. Luego se repite el protocolo de envio de datos de participantes explicado previamente pero ahora es el servidor quien envia los datos y el cliente quien los lee (es decir, 2 bytes indicando la longitud de los datos del participante y usando `;` como separador de los mismos). Vemos que se mantiene un protocolo coherente, no hay necesidad de cambiar la logica a la vuelta.

* Una vez enviados todos los datos de los ganadores del batch se procede a repetir el mismo procedimiento para el siguiente batch (cantidad de participantes en el batch y envio de datos para procesamiento) hasta que se procesa el ultimo batch del cliente.

* Cuando se termina de procesar el ultimo batch el cliente envia una longitud de 0 (es decir, los 2 bytes que indicarian la cantidad de participantes de un batch son tomados como flag de completitud si llegan en 0 al servidor). El servidor procede a escuchar conexiones para procesar a otro posible cliente.

Una vez finalizado el ultimo paso el servidor pasa a esperar por la conexion del proximo cliente.
