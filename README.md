# rtmpmate 1.x

> [[domain] http://studease.cn](http://studease.cn/rtmpmate.html)
> [[中文] https://blog.csdn.net/icysky1989/article/details/88946642](https://blog.csdn.net/icysky1989/article/details/88946642)

> 公众号：STUDEASE
> QQ群：528109813
> Skype: live:670292548

This is not only a modern Live Streaming Server, but also a WebSocket IM Framework and Distributed File Sync System. It has event-driven architecture, which removes a lot of potential flakiness.  

Supported codecs: 

- H264
- AAC

Inputs: 

- RTMP
- RTSP over TCP

Outputs: 

- RTMP
- RTSP over TCP
- HTTP/WS-FLV
- HTTP/WS-fMP4
- MPEG-DASH (CMAF)
- HLS (CMAF)

Main features: 

- Remote controllable FLV and fMP4 recording
- Live playback via HTTP/WS-FLV, MPEG-DASH and HLS
- IEventHandler module which supports reporting via HTTP or launching a process
- Clustering, Proxy, Load Balancing
- Original Protection
- WebSocket IM Framework
- Distributed File Sync System


## Roadmap

----------
### common

- Clone listener array while dispatching event

------------
### rtmpmate

- RTSP Live Streaming over TCP
- CMAF
- Register typed interfaces
   - IEventHandler
      - HttpEventHandler
      - ProcEventHandler
- Refactor
   - Move package "mux" and "dvr" into "av" (av.IMuxer, av.IDVR)
   - Remove rtmp.Stream. Implement dvr.Stream as av.IStream
      ```go
        type IRecordableStream interface {
            AddDVR(cfg *config.DVR) (dvr.IDVR, error)
            GetDVR(id string) dvr.IDVR
            RemoveDVR(id string)
        }

        type IStream interface {
            av.IReadableStream
            av.ISinkableStream
            IRecordableStream

            ReadyState() uint32
            Close() error
        }
      ```
   - Async event-driven handler for media packets
   - Bandwidth Manager
      ```go
        bw.Manager.Add(*bw.Reader)
      ```
   - Generate silent frame while the remote dropped some frames
- WebRTC Live Streaming

------------
### chatease

https://docs.oracle.com/cd/E19455-01/806-1387/6jam6926f/index.html

- publish/subscribe, message id
- access filter (login, visitor)
- message filter (frequency, rights, delivery, recorder)
- task queue on loading (register into name service) *
- hot update *

------------
### syncease

- sync on chunk

----------
### others

- Name Service
- Log Reporting
