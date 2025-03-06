# High Level View
We are not going to create a game spawning service yet.  Unless we get some how
so popular that i cannot run ~10,000s of users on a single cpu, then... well,
then i'll have to scale horizontally

## Why JSON?
I just didn't want to write parser 3 times.  so i am just using zig as the data
packet but the header is binary.  This will allow for fast intervention for
auth-proxy but simplicity.  Especially since i have no idea if i have the right
packets!

## topology
client([ ws conn ]) -> vimguys.theprimeagen.com
* this is for authentication
* reverse proxy for game to prevent game being overloaded

vimguys([ tcp conn]) -> internal.vimguys.game
* the ackshual game
* the game runner is going to have to create new games and resolve games

## Network Protocol
High level look at the protocol is the following:
[ version(2) | type(2) | len(2) | player_id(4) | game_id(4) | data(len) ]

### version
This represents two different types of versioning.

* binary protocal version.  This should rarely, if ever change.
* new `types`  This will be incremented and will be considered breaking every time.

**Update required** if out of sync
Should never change mid game

### player_id & game_id
this will be used by the game and the proxy.  They have no affect on the
clientside

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

type BillboardMessage = {
    title: MessageDisplay,
    display: MessageDisplay,
    time: number
}

type GameStatusMessage =
// displays a large central box
BillboardMessage & {
    type: 1,
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

// navigation markers
| {
    type: 3
    pos: Vec2
    keys: string
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
        display: BillboardMessage
        stats: Stats
    }
}

type MiniGameEnd = {
    win: boolean
    display: BillboardMessage
}

type MiniGameInit = {
    display: BillboardMessage
    map: {
        width: number,
        height: number,
        renders: RenderedObject[]
    }
}

type GameMessage = {
    data: MiniGameInit | MiniGameEnd
}

type Stats = { ... unknown at the time of writing ... }
```

## Decoding strategy
