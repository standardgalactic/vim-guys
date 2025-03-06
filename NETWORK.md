# High Level View
We are not going to create a game spawning service yet.  Unless we get some how
so popular that i cannot run ~10,000s of users on a single cpu, then... well,
then i'll have to scale horizontally

## topology
client([ ws conn ]) -> vimguys.theprimeagen.com
* this is for authentication
* reverse proxy for game to prevent game being overloaded

vimguys([ tcp conn]) -> internal.vimguys.game
* the ackshual game
* the game runner is going to have to create new games and resolve games

## Network Protocol
High level look at the protocol is the following:
[ version(1) | type(2) | len(2) | data(len) ]

### Authentication
```
Auth : <uuid>
AuthResponse : <bool> authenticated
```

### Connection Status / Maintenance
```
# types
Ping : <no data>
Pong : <no data>
ConnError : <str#len> connection error string
GameError : <str#len> game error string
```

### Rendering / Game
```
# types
...
```
