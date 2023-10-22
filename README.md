# ToyBT Client

This is a very simple implementation of a BitTorrent client written in Go. It was done following the outline provided by the CodeCrafters challenge. As it is, it can only download single file torrents and is not flexible (e.g. `bitfield` messages have to be received before proceeding, contrary to the more "lenient" protocol).

Below is a general outline of some of the implementation details and decisions made through the stages.

## Stages 1-4 - Bencode

The encoding/decoding was made from scratch. It covers barely enough to make the rest of the program functional. This means that it doesn't cover types in the best way (with type checking, reflection etc.), like specific libraries for bencode do. But it does cover the ones required for the rest of the challenge.

Parsing is mainly done using a simple recursive approach without complex pattern matching.

## Stages 6-8 - Torrent, Tracker

During this stages the main entities of the domain were created, namely Torrent, Tracker, Peer.

## Stages 9 - Networking

In this stages the `PeerConn` was implemented to provide a way to establish a connection with a peer. Each such object corresponded to one peer connection to download one piece at a time. As a future improvement, the connections to a peer could be kept to be reused for other pieces too.

In these stages only a basic handshake protocol is needed, so a connection is established and the handshake messages are exchanged.

## Stage 10 - Piece download

For a piece to be downloaded, the main structure of the connection handling had to be set up so it can asynchronously and concurrently handle events.

### Finite State Machine

The package `fsm` contains a very basic finite state machine implementation. It is used by `PeerConn` to keep the state and handle all the different cases that could arise, since the communication will be asynchronous. The output message after each transition is applied will be routed to trigger the appropriate event handler in the `handleEventQueue` routine.

### Event Handling

Every event is placed in a channel (`eventQueue`). This channel is being monitored by the routine mentioned above. The event name will be passed to FSM to apply the transformation, change its state and get the output message. This message dictates the handler that will run. The `have_bitfield` and `interested` cases are run in the same routine, whlie for the `request` and `save_piece` cases, new goroutines are spawned. This was chosen because the latter will pipeline their events to speed up the downloading, meaning the event handler should be free to handle the next events that come.

### Listening to incoming messages

The `listen` function constantly listens for new messages from the peer and adds a corresponding event to the queue.

### Writing messages to peer

The `write` function is used to send messages to the peer. A mutex lock is used to avoid race conditions.

### `interested` handler

The `produceInterested` handler function initializes the piece object that will be held by the `PeerConn` object and used to request and save data.

### `request` handler

The `produceRequest` handler function splits the piece into blocks and pipelines request messages for them. `q` is used to limit the number of pipelined requests. The errors that may occur during these requests are piped to the "main" routine (which in this case is the event handling routine) for processing via the `errChan` channel.

### `save_piece` handler

The `handlePiece` handler function writes the block received to the buffer held by the piece currently downloading. When the piece is complete, its integrity is verified against the hash provided in the torrent file. After that, the buffer is written to storage. For more flexible implementations, storage is of `io.Writer` type. In this stages it will actually be a file. In the next stage it will be a buffer held in memory.

The `signal` channel is used for synchronizing the caller routine with the inner concurrent routines running to download the piece.

## Stage 11 - Downloading a file

### File download service

The `DownloadFileService` first initializes buffers in memory to hold the pieces and then concurrently initiates pieces downloads using a channel that holds the pieces-tasks. If an error is encountered during a download, the piece is put back in the queue. In the main thread, a counter is kept to know when all the pieces have been download. After that, the pieces are written to a file.

## Next Steps / Possible Improvements

- Support for torrents with multiple files
- Wider and more lenient protocol implementation
- CLI improvements
- Peer connection reuse for multiple piecees

