# raft
This repo shows how to implement a simple raft. The environment is designed for course SJTU NIS3607 project-1.
## Running
代码模板地址：![Github](https://github.com/Waynezee/SimpleConsensus)
```shell
mylogger: 实现简单的日志记录功能
myrpc: 实现简单的rpc功能，节点间通过rpc进行相互通信
core: 待实现的共识协议
config: 保存节点的配置文件
log: 保存节点的日志信息
cmd: 保存可执行文件
scripts: 测试和运行脚本
```

- 测试流程：
```shell
git clone https://github.com/Waynezee/SimpleConsensus
cd SimpleConsensus/ && mkdir log
cd cmd/ && make
cd ../scripts/ && chmod +x *.sh
./run_nodes.sh [test time (seconds)]
等待运行结束
python3 check.py 7
当编写你的代码的时候需要注意：
需要保证cmd/node.go的输入参数形式与原始的一致
保证core/block.go中getBlock()和commitBlock()函数不变，并在合适的时候进行调用
getBlock(): 生成一个区块，同时会将相关信息记录到日志中
commitBlock() : 当协议确定性的完成共识后进行调用，同时会将相关信息记录到日志中
```
