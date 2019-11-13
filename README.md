# 协议问题记录

## packetid有序,起始packetid编号如何定

若从零开始,则存在问题: 客户端可能重启,发送编号seq被初始化; 此时服务端已接收编号recvd_seq比客户端编号大,不接收.

### 解决方案

1. 客户端建立连接时发送序号,接收服务端响应已接收编号
2. 服务端检测客户端源端口,若端口号变,则已接收编号归零
3. 服务端2s内未聚合完分片则丢弃.当所有分片被丢弃时,已接收编号归零.重启的客户端间隔2s即可重启编号
   1. merge fragment, ready packet
   2. cleartimeout fragments and packets

refresh min_readyid
    timeouted, 可能增加,清空时置位0
    new packet frame recved

首先接收上层数据,合并数据,丢弃过时数据
然后下层需要接收, 发送查找信号,接收数据
