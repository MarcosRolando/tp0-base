Para este ejercicio, el servidor se encuentra esperando conexiones de clientes que enviaran los datos de los participantes de su dataset correspondiente para luego contestarles con los datos de los participantes ganadores. Esto es procesado en forma de batch. Ademas, al finalizar el envio de los datos en el cliente, el mismo consulta el numero total de ganadores al servidor contabilizando todas las agencias. En caso de que el servidor este procesando el batch de algun cliente en el momento de la request se le respondera al cliente con el resultado parcial de ganadores que se tiene registrado en ese momento junto con el numero total de agencias cuyos batches esten siendo procesados en ese momento. Si no hay agencias siendo procesadas entonces se devuelve unicamente el total de ganadores que se tiene registrado y el cliente termina su ejecucion, sino reintentara cada N segundos la misma request.

Todos los datos numericos enviados en el protocolo son en Big Endian. Cualquier error en el protocolo corta la comunicacion del servidor con el cliente y no hay reintentos. Cada request de procesamiento de participantes en batch o request de los resultados de los ganadores totales se realiza en una nueva conexion (a diferencia del ejercicio anterior, donde todos los batches se procesaban en serie en un unico servidor). De esta forma es mas fair con el resto de los clientes ya que no deben esperar tanto para procesar sus datos en caso de que haya clientes con datasets mucho mas grandes.

El protocolo utilizado en este caso es el siguiente:

* Cuando un cliente se conecta al servidor primero envia un unico byte que indica el tipo de request: 0 si se va a enviar un batch de participantes o 1 si se quiere obtener el total de ganadores registrados. 

* Si es un request de tipo batch el protocolo de envio de participantes y respuesta del servidor es el mismo que en el `ej6` y `ej7`, referirse al `PROTOCOL.md` del `ej6`.

* Una vez procesado el batch el servidor procede a escuchar por nuevas requests de clientes (previamente solo realizaba esto cuando el cliente enviaba todos los batches, lo que restringia el acceso al servidor de parte de otros potenciales clientes).

* Si es un request de tipo resultados de ganadores el servidor respondera primero con un unico byte que indica el tipo de respuesta: 0 si es un resultado parcial o 1 si es un resultado total. En el caso del resultado parcial el servidor envia luego 2 bytes indicando el numero de agencias siendo todavia procesadas seguido de 4 bytes que indican el numero de ganadores registrados hasta ese momento. Si el resultado es total unicamente se envian los 4 bytes con el numero de ganadores.

Es importante aclarar que debido a que es a priori imposible saber cuantas agencias totales existen o existiran se considera como un tipo de respuesta total aquella en la que al momento de realizar la consulta la central no este procesando a ninguna agencia. Podria igual suceder que despues de esto se procese a alguna nueva agencia que recien comience a enviar sus datos, por lo que el resultado total de ganadores que recibe una agencia no implica el resultado final necesariamente.

## Sincronizacion

Respecto a los mecanismos de sincronizacion, para el caso de loggeo de los ganadores al archivo se mantiene el Lock utilizado en el ejercicio 7 (ver `PROTOCOL.md` de la branch `ej7` para referencia).

Para llevar el conteo total de ganadores entre todos las instancias del servidor se utilizo el mecanismo `Value` provisto por el package `multiprocessing`, que provee mecanismos para evitar las race conditions al manipularlo.

Para distinguir si algun servidor esta procesando o no el batch de algun cliente, se utilizo el mecanismo `Array` provisto por el package `multiprocessing`. Es similar a un `Value` pero la data almacenada es un array donde cada proceso tiene asignada una correspondiente entrada en el mismo. Se tratan de entradas binarias donde el valor 0 indica que no se esta procesando ningun batch mientras que la entrada 1 indica que si se esta procesando un batch. El unico proceso que puede modificar una entrada es el proceso al que le corresponda la misma. Luego cuando un cliente consulta por el resultado total de ganadores el proceso que recibe dicha consulta verifica que todas esas entradas esten en 0 para considerar la respuesta como total, caso contrario responde con una respuesta parcial donde ademas puede saber cuantas agencias estan procesando clientes simplemente con contar cuantos valores del Array se encuentran seteados en 1.


