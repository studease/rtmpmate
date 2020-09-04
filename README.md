# rtmpmate 1.x

> [[domain] http://studease.cn](http://studease.cn/rtmpmate.html)  
> [[中文] https://blog.csdn.net/icysky1989/article/details/88946642](https://blog.csdn.net/icysky1989/article/details/88946642)  
> 公众号：STUDEASE  
> QQ群：528109813  
> Skype: live:670292548  

This is not only a modern Live Streaming Server, but also a WebSocket IM Framework and Distributed File Sync System. It has event-driven architecture, which removes a lot of potential flakiness.

## Roadmap

- [ ] Async handler for media packets.
- [ ] Bandwidth Manager.

  ```go
  bw.Manager.Add(*bw.Reader)
  ```

### rtmpmate

- Codecs
  - [x] H264
  - [x] AAC

- Inputs
  - [x] RTMP
  - [ ] RTSP over TCP

- Outputs
  - [x] RTMP
  - [ ] RTSP over TCP
  - [ ] HTTP/WS-FLV
  - [ ] HTTP/WS-fMP4
  - [ ] MPEG-DASH (CMAF)
  - [ ] HLS (CMAF)
  - [ ] WebRTC

- Features
  - Backend API.
    - [ ] Remote control interface (RCI) of FLV and fMP4 DVR.
    - [ ] RTMP stats of server and application.
  - [ ] Cross-Origin Resource Sharing (CORS).
  - [ ] Cluster, Proxy, Load Balance.
  - IEventHandler module.
    - [ ] HttpHandler
    - [ ] ProcHandler

### chatease

WebSocket IM Framework (publish/subscribe, message id).

https://docs.oracle.com/cd/E19455-01/806-1387/6jam6926f/index.html

- [ ] Access filter (login, visitor).
- [ ] Message phase (frequency, rights, delivery, recorder).
- [ ] Task queue on loading (register into name service).
- [ ] Hot update.

### syncease

Distributed File Sync System.

- [ ] Sync on chunk.

### others

- [ ] ~~GSLB~~
- [ ] Name service
- [ ] User system
- [ ] Log reporting
