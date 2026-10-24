[English](../en/realtime.md) · **🌐 Español**

# Tiempo real

## 1. Componentes

```mermaid
flowchart LR
    subgraph Instance["Instancia del backend"]
        direction TB
        S[ServeWs] --> C1[Cliente A<br/>readPump · writePump]
        S --> C2[Cliente B]
        H[Hub<br/>bucle de eventos]
        C1 --- H
        C2 --- H
    end
    H <--> BUS[(Canal de Redis<br/>webchat:ws:events)]
    S --> T[(Redis<br/>tickets)]
    S --> PR[(Redis<br/>presencia)]
```

- **Hub:** un único bucle de eventos que sabe qué clientes pertenecen a cada sala.
- **Cliente:** uno por conexión, con una goroutine de lectura y otra de escritura.
- **Bus:** publica cada evento en Redis para que todas las instancias lo entreguen a *sus* clientes.
- **Presencia:** un sorted set de Redis con `userID:clientID` puntuado por el último latido. Un usuario está en línea si al menos una conexión latió en los últimos 2 minutos.

## 2. Conexión

```mermaid
sequenceDiagram
    participant B as Navegador
    participant A as API
    participant R as Redis
    B->>A: POST /ws-ticket (Bearer)
    A-->>B: ticket (30 s, un solo uso)
    B->>A: GET /ws?ticket=…
    A->>R: canjear ticket (GETDEL)
    A->>A: cargar las salas del usuario (RoomLister)
    A->>A: suscribir al cliente a cada sala
    A->>R: presencia en línea → difundir a las salas
    A-->>B: conectado
```

## 3. Envío de un mensaje

```mermaid
sequenceDiagram
    participant A as Alicia (instancia 1)
    participant I1 as Backend 1
    participant DB as PostgreSQL
    participant R as Bus Redis
    participant I2 as Backend 2
    participant B as Beto (instancia 2)
    A->>I1: POST /messages
    I1->>DB: INSERT
    I1->>R: publicar {kind: broadcast, type: new_message, sala}
    R-->>I1: evento
    R-->>I2: evento
    I1-->>A: new_message
    I2-->>B: new_message
```

Quien envía también recibe el evento, y así se confirma el mensaje optimista.

## 4. Eventos

Servidor → cliente, con `type` y `payload`:

| Tipo | Se envía cuando |
|---|---|
| `new_message` | se crea un mensaje |
| `update_message` | se edita un mensaje |
| `delete_message` | se borra un mensaje (el autor o un moderador) |
| `pin_message` | se fija o desfija un mensaje |
| `reaction_added`, `reaction_removed` | cambian las reacciones |
| `role_changed`, `member_removed` | cambia el rol de un miembro o se le expulsa |
| `typing` | un miembro está escribiendo (efímero, nunca se guarda) |
| `presence` | un miembro se conecta o desconecta |

Cliente → servidor: solo `{ "type": "typing", "room_id": … }`, aceptado de miembros de la sala y con límite de frecuencia.

Entre instancias el bus transporta cuatro clases: `broadcast`, `subscribe`, `unsubscribe` y `disconnect`. Las tres últimas aplican un cambio de membresía (entrar a una sala, ser expulsado, borrar la cuenta) en la instancia que tenga la conexión del usuario.

## 5. Modelo de fiabilidad

- **Entrega como máximo una vez.** Si un cliente está desconectado, pierde eventos. Es un compromiso deliberado frente al coste de un registro de eventos (ver D8 en [Decisiones de diseño](decisions.md)).
- **Recuperación en el cliente.** `SocketContext` reconecta con backoff exponencial (`lib/backoff.ts`) y pide un ticket nuevo cada vez. Tras reconectar, `useMessages` y `useRoomActivity` recargan los últimos mensajes y miembros, de modo que la pantalla converge al estado real.
- **Fuente de verdad.** La base de datos. Los eventos son avisos que hacen la interfaz rápida; una recarga siempre la repara.

## 6. Escalar horizontalmente

```mermaid
flowchart TB
    LB[Balanceador] --> I1[Backend 1]
    LB --> I2[Backend 2]
    LB --> I3[Backend N]
    I1 & I2 & I3 <--> R[(Redis pub/sub + presencia)]
```

No se necesitan sesiones pegajosas: un cliente puede conectarse a cualquier instancia y seguir viendo los eventos que se originan en las demás. La planificación de capacidad y las pruebas de carga (k6) están en el roadmap.
