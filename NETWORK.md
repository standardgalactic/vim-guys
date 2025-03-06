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
[ version(1) | type(2) | len(2) | player_id(4) | data(len) ]

### Notes
the `player_id` will be filled in from the auth.  if any value is within that
spot, it will be overwritten.

### Authentication
```typescript
type Auth = {
    data: UUID
}

type AuthResponse = {
    data: boolean
}
```

### Connection Status / Maintenance
```typescript
type Ping = {
    data: null
}

type Pong = {
    data: null
}

type ConnectionTimings = {
    data: {
        sent: number,
        received?: number
    }
}

type ConnError = {
    data: string
}

type MessageDisplay = {
    msg: string,
    foreground: Hex
    background?: Hex
    bold?: boolean // i think vim supports this
}

type GameStatusMessage =
// displays a large central box
{
    type: 1,
    title: MessageDisplay,
    display: MessageDisplay,
    time: number
}

// displays a game message in bottom row
| {
    type: 2,
    display: MessageDisplay,
    time: number
}

type GameStatus = {
    data: GameStatusMessage
}

type SystemMessage = {
    data: string
}

```

### Rendering
**Needs**
Some assets, players, will move in non predictable patterns
Some assets, items, could visible until someone picks one up
Some assets, walls, never move
Some assets, moving walls, could move with a cycle

```typescript
type RenderedObject =
// this should cover 1 - 3
{
    type: 1,
    id: number,
    pos: Vec2
    rect: [Vec2, Vec2] // n x m
    color:
        number | // all one color
        number[] // length = n x m
}

// this should cover number 2
| {
    type: 2,
    startingPos: [Vec2, Vec2] // n x m
    endingPos: [Vec2, Vec2] // n x m
    timings: {
        cycleTime: number,
        offsetTime: number,
        cycleDelay: number,
    }
    rect: [Vec2, Vec2] // n x m
    color:
        number | // all one color
        number[] // length = n x m
}

type Rendered = {
    data: RenderedObject
}
```

### Input
**Context**
Now with vim we do not get to know if two keys are held down at once.  But we
have vim navigation which has been designed around the fact you cannot do that.
We can also not know if two keys are held down at the same time.  So this does
limit us in the natural gaming sense.

```typescript
type Input = {
    data: { key: string }
}
```

### Game Information
```typescript
type PlayerId = {
    data: number
}
type GameCountdown = {
    data: number
}
type GameOver = {
    data: {
        win: boolean // i am sure i will want more information here
        stats: { ... } // i'll have to think about
    }
}
```
