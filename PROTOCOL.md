El protocolo utilizado es el mismo que en el `ej6`, refierase al `PROTOCOL.md` correspondiente de ese branch para una descripcion del mismo.

Para el manejo de las conexiones en paralelo, se realiza un fork del proceso original el cual crea el socket aceptador. Los procesos hijos luego utilizan las copias de este socket para ir aceptando conexiones de clientes.

La sincronizacion del acceso al archivo se realiza mediante un Lock provisto por el package `multiprocessing`, de forma tal que ningun proceso pueda escribir al archivo en simultaneo con otro. Dicho lock solo es adquirido en el momento del loggeo, logicamente, para maximizar el paralelismo.
