# AirIoT Driver List

Complete mapping table of AirIoT driver names and identifiers.

## Driver List

| 驱动名称 | Driver ID | Install Command | 描述 |
|---------|-----------|-----------------|------|
| AB PLC | `abplc` | `npx kesi abplc` | Allen-Brandley PLC 驱动，支持AB PLC系列设备接入 |
| AB 小型PLC | `abplc_mic` | `npx kesi abplc_mic` | Allen-Brandley 小型 PLC 驱动，支持MicroLogix系列 |
| 艾森Lora/A11 GRM | `aisenz_a11` | `npx kesi aisenz_a11` | 艾森 Lora A11 GRM 设备驱动，支持无线Lora通信 |
| 艾森Lora/TDM | `aisenz_tdm` | `npx kesi aisenz_tdm` | 艾森 Lora TDM 设备驱动，支持时分复用通信 |
| BACnet驱动 | `bacnet` | `npx kesi bacnet` | BACnet 楼宇自动化协议驱动，支持暖通空调设备 |
| 数据接口驱动 | `data-service-driver` | `npx kesi data-service-driver` | 数据接口服务驱动，通过API接口采集数据 |
| DB（数据库）驱动 | `db-driver` | `npx kesi db-driver` | 数据库采集驱动，支持MySQL/Postgres/Oracle/Sqlserver/达梦/MONGO |
| driver-camera-gb28181 | `driver-camera-gb28181` | `npx kesi driver-camera-gb28181` | GB28181 摄像头驱动，支持国标视频设备接入 |
| driver-charging-kjtc | `driver-charging-kjtc` | `npx kesi driver-charging-kjtc` | KJTC 充电桩驱动，支持电动汽车充电桩设备 |
| CoAP服务端 | `driver-coap` | `npx kesi driver-coap` | CoAP 物联网协议服务端驱动，支持资源受限设备 |
| driver-common-camera | `driver-common-camera` | `npx kesi driver-common-camera` | 通用摄像头驱动，支持RTSP协议摄像头 |
| driver-ctwing-lwm2m | `driver-ctwing-lwm2m` | `npx kesi driver-ctwing-lwm2m` | 天翼 LwM2M 物联网驱动，支持电信NB-IoT设备 |
| driver-ctwing-mq | `driver-ctwing-mq` | `npx kesi driver-ctwing-mq` | 天翼 MQ 消息队列驱动，支持电信物联网平台 |
| 萤石云驱动 | `driver-ezviz` | `npx kesi driver-ezviz` | 萤石云摄像头驱动，支持海康萤石云平台视频预览回放云台控制 |
| driver-gato-alarm | `driver-gato-alarm` | `npx kesi driver-gato-alarm` | GATO 报警设备驱动，支持报警主机接入 |
| driver-gw-szyh | `driver-gw-szyh` | `npx kesi driver-gw-szyh` | 深圳远怀网关驱动，支持物联网网关设备 |
| 海康报警主机 | `driver-hik-alarm` | `npx kesi driver-hik-alarm` | 海康威视报警主机驱动，支持报警主机接入和布防撤防 |
| 海康摄像头驱动 | `driver-hik-camera` | `npx kesi driver-hik-camera` | 海康威视摄像头驱动，支持预览云台控制接收报警 |
| 海康摄像头驱动（旧） | `driver-hik-camera-direct` | `npx kesi driver-hik-camera-direct` | 海康威视摄像头驱动(旧版)，保留兼容性 |
| driver-hik-isc-camera | `driver-hik-isc-camera` | `npx kesi driver-hik-isc-camera` | 海康 ISC 平台摄像头驱动，支持ISC平台接入 |
| driver-hj212 | `driver-hj212` | `npx kesi driver-hj212` | HJ212 环保数据传输协议驱动，支持环保监测设备 |
| driver-http-client | `driver-http-client` | `npx kesi driver-http-client` | HTTP 客户端驱动，通过HTTP请求采集数据 |
| driver-icpas-spon | `driver-icpas-spon` | `npx kesi driver-icpas-spon` | ICPAS SPON 设备驱动，支持SPON系列设备 |
| driver-java-001 | `driver-java-001` | `npx kesi driver-java-001` | Java 自定义驱动示例，用于驱动开发参考 |
| driver-jt100-nova | `driver-jt100-nova` | `npx kesi driver-jt100-nova` | JT100 Nova 车载终端驱动，支持车辆定位终端 |
| driver-ky-NB-mqtt | `driver-ky-NB-mqtt` | `npx kesi driver-ky-NB-mqtt` | 科远 NB-IoT MQTT 驱动，支持科远NB设备 |
| driver-ky-http-mqtt | `driver-ky-http-mqtt` | `npx kesi driver-ky-http-mqtt` | 科远 HTTP MQTT 驱动，支持科远网关设备 |
| driver-lamp-sz | `driver-lamp-sz` | `npx kesi driver-lamp-sz` | 深圳路灯控制驱动，支持智能路灯系统 |
| driver-led-lednets | `driver-led-lednets` | `npx kesi driver-led-lednets` | LED 灯联网控制驱动，支持LED显示屏控制 |
| LwM2M服务端 | `driver-lwm2m-server` | `npx kesi driver-lwm2m-server` | LwM2M 服务端驱动，轻量级物联网设备管理协议 |
| GB28181服务器驱动 | `driver-media-server` | `npx kesi driver-media-server` | GB28181 媒体服务器驱动，支持国标视频设备注册接入 |
| MQTT驱动 | `driver-mqtt-client` | `npx kesi driver-mqtt-client` | MQTT 客户端驱动，支持订阅发布消息模式采集 |
| driver-nn-ctwing-nb-ddkzq | `driver-nn-ctwing-nb-ddkzq` | `npx kesi driver-nn-ctwing-nb-ddkzq` | 南宁电信 NB-IoT 驱动，支持电信NB设备 |
| driver-nn-ctwing-nb-ddkzq-g340t | `driver-nn-ctwing-nb-ddkzq-g340t` | `npx kesi driver-nn-ctwing-nb-ddkzq-g340t` | 南宁电信 NB-IoT G340T 驱动，支持G340T模组 |
| OneNet MQTT | `driver-onenet-mqtt` | `npx kesi driver-onenet-mqtt` | OneNet MQTT 平台驱动，支持中移物联OneNet平台 |
| driver-radio-gxh | `driver-radio-gxh` | `npx kesi driver-radio-gxh` | 广西信呼无线电设备驱动，支持无线电通信设备 |
| driver-radio-shibang | `driver-radio-shibang` | `npx kesi driver-radio-shibang` | 世邦无线电设备驱动，支持无线对讲设备 |
| driver-shunzhouReport-nn | `driver-shunzhouReport-nn` | `npx kesi driver-shunzhouReport-nn` | 顺舟南宁上报驱动，支持顺舟物联网设备 |
| SNMP网络管理站 | `driver-snmp` | `npx kesi driver-snmp` | SNMP 网络管理协议驱动，支持网络设备监控 |
| driver-telecomReport-nn | `driver-telecomReport-nn` | `npx kesi driver-telecomReport-nn` | 电信南宁上报驱动，支持电信物联网平台 |
| 拓天周界驱动 | `driver-ttzj` | `npx kesi driver-ttzj` | 拓天周界报警驱动，支持周界防范系统 |
| driver-wifi-kx | `driver-wifi-kx` | `npx kesi driver-wifi-kx` | WiFi 科远设备驱动，支持WiFi物联网设备 |
| driver-zl-ctwing | `driver-zl-ctwing` | `npx kesi driver-zl-ctwing` | 中联电信物联网驱动，支持电信物联网设备 |
| driver-zldzc | `driver-zldzc` | `npx kesi driver-zldzc` | 中联电子站牌驱动，支持智能公交站牌 |
| EU63驱动 | `eu63` | `npx kesi eu63` | EU63 协议驱动，支持EU63通信协议设备 |
| G340T | `g340t` | `npx kesi g340t` | G340T NB-IoT 模组驱动，支持移远G340T模组 |
| 海康报警主机SDK（旧） | `hik-alarm-sdk` | `npx kesi hik-alarm-sdk` | 海康威视报警主机 SDK 驱动(旧版)，保留兼容性 |
| 海康门禁SDK驱动 | `hikdoor-sdk-driver` | `npx kesi hikdoor-sdk-driver` | 海康威视门禁 SDK 驱动，支持门禁控制设备 |
| HTTP服务器驱动 | `http-server-driver` | `npx kesi http-server-driver` | HTTP 服务器驱动，接收HTTP请求上报数据 |
| IEC104电力规约 | `iec104` | `npx kesi iec104` | IEC 104 电力规约驱动，支持电力行业标准通信 |
| IEC61850 MMS v3 | `iec61850-mms-v3` | `npx kesi iec61850-mms-v3` | IEC 61850 MMS v3 电力驱动，支持变电站自动化 |
| IEC61850 MMS v4 | `iec61850-mms-driver` | `npx kesi iec61850-mms-driver` | IEC 61850 MMS v4 电力驱动，支持电力设备通信 |
| Kafka驱动 | `kafka-driver` | `npx kesi kafka-driver` | Kafka 消息队列驱动，支持Kafka消费者模式 |
| 三菱MC驱动 | `mitsubishi-mc` | `npx kesi mitsubishi-mc` | 三菱 MC 系列 PLC 驱动，使用MC协议以太网通信 |
| Modbus驱动 | `modbus` | `npx kesi modbus` | Modbus RTU/TCP 协议驱动，支持01-04功能码读写 |
| 国传Modbus变频器 | `modbus-vfd` | `npx kesi modbus-vfd` | 国传 Modbus 变频器驱动，支持变频器设备 |
| Modbus RTU驱动 | `modbus_rtu` | `npx kesi modbus_rtu` | Modbus RTU 协议驱动，支持串口Modbus通信 |
| 欧姆龙PLC | `omron` | `npx kesi omron` | 欧姆龙 PLC 驱动，支持FINS协议以太网通信 |
| OneNet平台驱动 | `onenet` | `npx kesi onenenet` | 中移物联 OneNet 平台驱动，支持OneNet老版平台 |
| OPCDA驱动 | `opcda` | `npx kesi opcda` | OPC DA 自动化接口驱动，基于Windows COM/DCOM技术 |
| OPCUA驱动 | `opcua` | `npx kesi opcua` | OPC UA 统一架构协议驱动，支持订阅/直接读取模式 |
| 西门子S7驱动 | `siemens-s7` | `npx kesi siemens-s7` | 西门子 S7 系列 PLC 驱动，支持S7-300/1200/1500 |
| 西门子S7-200驱动 | `siemens-s7-200s` | `npx kesi siemens-s7-200s` | 西门子 S7-200 SMART PLC 驱动，支持S7-200系列 |
| 水利651-2014 | `sl651-2014-driver` | `npx kesi sl651-2014-driver` | 水利 651-2014 协议驱动，支持水文监测数据传输 |
| TCP客户端驱动 | `tcp-client-driver` | `npx kesi tcp-client-driver` | TCP 客户端驱动，主动连接TCP服务器采集数据 |
| TCP服务器驱动 | `tcp-server-driver` | `npx kesi tcp-server-driver` | TCP 服务器驱动，接收TCP客户端连接上报 |
| TCP漏检驱动 | `tcp-server-driver-leak-detection` | `npx kesi tcp-server-driver-leak-detection` | TCP 服务器漏检驱动，支持漏水检测系统 |
| TLS350消防驱动 | `tls350-driver` | `npx kesi tls350-driver` | TLS350 消防设备驱动，支持消防报警系统 |
| UDP客户端驱动 | `udp-client-driver` | `npx kesi udp-client-driver` | UDP 客户端驱动，主动发送UDP请求采集数据 |
| UDP服务器驱动 | `udp-server-driver` | `npx kesi udp-server-driver` | UDP 服务器驱动，接收UDP数据包并解析 |
| WebSocket客户端驱动 | `websocket-client-driver` | `npx kesi websocket-client-driver` | WebSocket 客户端驱动，主动连接WS服务器 |
| WebSocket服务器驱动 | `websocket-server-driver` | `npx kesi websocket-server-driver` | WebSocket 服务器驱动，接收WS客户端连接 |
