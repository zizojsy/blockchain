# 构造自己的区块链

## 使用方法
```
Usage:
  wallet
    -c
      Create a new account in wallet
    -l
      List all accounts in wallet
    -T -f A -t B -a AMOUNT [-m]
      Transfer AMOUNT money from A to B, mine coin if -m flag is set

  service
    -s [-m ADDRESS]
      Start service, mine coin if ADDRESS is given
    -p
      Print all blocks in the blockchain
    -b ADDRESS
      Get balance of ADDRESS
```

## P2P多终端设定(Windows PowerShell)
*注：所有终端在第一次运行任意指令时都会先创建**创世块**，并将其保存在数据库中。*
```
$env:node_id=[端口号]
```
### 示例
1. <a name="fullnode">全节点</a>：本节点理论上模拟创建**创世块**的创建者，由于本代码**创世块**的建立是在代码中硬编码的(与*BitCoin*类似)，而**传世块**需要创建一个交易产生最初的资金来源，故本代码简化逻辑，本节点的账号也是硬编码，并且不能创建账号。

    ```
    $env:node_id=3000
    ```

2. 矿工节点：本节点负责挖矿，启动服务时需要指定收益账号
    ```
    $env:node_id=3010
    ```

3. SPV:交易验证节点
    ```
    $env:node_id=3020
    ```

## 实操
### 1. 创建钱包，并将钱包持久化至本地。
```
    $> blockchain wallet -c
```
*结果*
```
    Done!
    Your new address: 1DMm8boViwyVMyts9ce6pFxikLqHv3VSa4 
```
注：全节点不能创建钱包，具体原因参见 [P2P示例中关于全节点描述](#fullnode)

### 2. 显示当前终端所持有的账户地址
```
    $> blockchain wallet -l
```
*结果*
```
    12YdVGoSepFPea677pGjWpUgcY3eCMNxX8
    1DMm8boViwyVMyts9ce6pFxikLqHv3VSa4
    17iMWJo9nuUXPc4MU4sKrWQEScEJ62vrgd
    1Kv1cjv1dXVnpNNj39CCd71vmDfGyqDWig 
```

### 3. 转账
 **-f** 转出地址<br>
 **-t** 接受地址<br>
 **-a** 转账金额<br>
 **-m** (可选)，参与挖矿<br>

```
    $> blockchain wallet -T -f 187wiR3JP6bEBewTkkTcSZNfCRmwTyteYz -t 15VSra4M24knbrpAfeqSfVEeyY8Qag4GLr -a 10 -m
```
*结果*
```
    Done!
    746cc40b5f7b1aa489b7e3c7815c7dc70571689515674c0857a7a9f847494fe4

    Success!
```

### 4. 启动服务，同步数据并保存至数据库中
```
    $> blockchain service -s -m 1DMm8boViwyVMyts9ce6pFxikLqHv3VSa4
```
 **-s** 启动服务<br>
 **-m** (可选)，参与挖矿，指定奖金接受地址
*结果*
```
    Starting node 3000
    Received version command
    Received getblocks command
    Received getdata command
    Received getdata command
    Received getdata command
```

### 5. 打印区块链
```
    $> blockchain service -p
```
*结果*
```
============ Block 0000055eb8192823c0942271b7b4fb2f45f449bec474db57e0bc8b9dd8f5bd33 ============
Height: 4
Prev. block: 0000206d43c9dc4912b7a13e68aed42fd844f406549b2cd0cd8dbb0d40dfd7a9
PoW: true

--- Transaction f9ea5d63977fda418029df39768b2d2009c53c419124545bfe46cdcd2be91529:
     Input 0:
       TXID:
       Out:       -1
       Signature:
       PubKey:    65663165366265646264343935636664363430613565356431646130373836356437356666306232
     Output 0:
       Value:  10
       Script: 4e190c9afd4c7bcb1f09e8263a26ea49e49ced31
--- Transaction a345a701d9102d19f49de0f03e5c5769b2c409269a8059e4eae9d42432d4f170:
     Input 0:
       TXID:      66f6bee4ec0891ff11b8b1924227cbfc8e687f13d727baa15f158efcd6bf1506
       Out:       0
       Signature: ef4c4152b032fca2cdcc73a28d030065ba9e1c8c5f741fabb9de1b80673749bd526517a52f7973c92574f7067602cb7cd75f78eadc104a351b96424a30527319
       PubKey:    28f46a4493c79fe6f0a9475de342464b7cc27eae9e259302d925c0f26daf9aa91e07735a8d36ade16c043cd22f1216a91508ecfa38e44a73ec38ba77a394b76e
     Output 0:
       Value:  10
       Script: 3141bb817159aafa64161968f26f162b2bef4b4a


============ Block 0000206d43c9dc4912b7a13e68aed42fd844f406549b2cd0cd8dbb0d40dfd7a9 ============
Height: 3
Prev. block: 00008fa7386c51feed3f8fbba5b7beb910fad16b440ab54705288287c7626fca
PoW: true

--- Transaction 66f6bee4ec0891ff11b8b1924227cbfc8e687f13d727baa15f158efcd6bf1506:
     Input 0:
       TXID:
       Out:       -1
       Signature:
       PubKey:    61313363623435323738653436316666356161643532376135383062386336363665613361633963
     Output 0:
       Value:  10
       Script: 4e190c9afd4c7bcb1f09e8263a26ea49e49ced31
--- Transaction e774be66ef398cfb6417c46582a3abf12f764b3c641dba249f6382146ed4f687:
     Input 0:
       TXID:      52b5b513410f76d1676b027d8b84d6da66d186872b222ae88ba303d054c6232a
       Out:       0
       Signature: d92f4b7edaba2f2e64b9a4e7d5485b3d2aa5307d0dca1b56e6c969668432bfab72552526b48f7d405435156af430f1b1a77ebd82e1832da08de7a13b3636815b
       PubKey:    28f46a4493c79fe6f0a9475de342464b7cc27eae9e259302d925c0f26daf9aa91e07735a8d36ade16c043cd22f1216a91508ecfa38e44a73ec38ba77a394b76e
     Output 0:
       Value:  10
       Script: 3141bb817159aafa64161968f26f162b2bef4b4a


============ Block 00008fa7386c51feed3f8fbba5b7beb910fad16b440ab54705288287c7626fca ============
Height: 2
Prev. block: 00002d105145906960d6d456fdd6f95fa005e065d864f57a478785d6f9050953
PoW: true

--- Transaction 52b5b513410f76d1676b027d8b84d6da66d186872b222ae88ba303d054c6232a:
     Input 0:
       TXID:
       Out:       -1
       Signature:
       PubKey:    35623630623930323239653030363638653631343035323433633335393233373730643337663537
     Output 0:
       Value:  10
       Script: 4e190c9afd4c7bcb1f09e8263a26ea49e49ced31
--- Transaction bbcb72ee08c8d99aa72bb1c2e78102a25a8681272ec17cedd51efdf7a4566140:
     Input 0:
       TXID:      9712f116cdb0dde40d2ca285c7dfcabe59a57b56b7472aed452c17821af97792
       Out:       0
       Signature: cff43e33b15b1bf1c59fe59a7b0b2608a0fb1591c4d47f88d733d5b1001f668be1d9f2553b5dc826a4feac34b6a0e7aac5b6f01864c44592acb5dd89c282a841
       PubKey:    28f46a4493c79fe6f0a9475de342464b7cc27eae9e259302d925c0f26daf9aa91e07735a8d36ade16c043cd22f1216a91508ecfa38e44a73ec38ba77a394b76e
     Output 0:
       Value:  10
       Script: 3141bb817159aafa64161968f26f162b2bef4b4a


============ Block 00002d105145906960d6d456fdd6f95fa005e065d864f57a478785d6f9050953 ============
Height: 1
Prev. block: 00002b40928246808d0b7fe186b63ddb36659e047800b76012d65327be8adad0
PoW: false

--- Transaction 9712f116cdb0dde40d2ca285c7dfcabe59a57b56b7472aed452c17821af97792:
     Input 0:
       TXID:
       Out:       -1
       Signature:
       PubKey:    65643936616664393936363362626434373965613234323738396237376336343031323136386337
     Output 0:
       Value:  10
       Script: 4e190c9afd4c7bcb1f09e8263a26ea49e49ced31
--- Transaction 5a50bb606a178e22919b8b47eee4860fdca3ec9397b5471a5a6eefd021f3dbdc:
     Input 0:
       TXID:      a7b3ddf2cf5658d10d126a73ee74d7b84a04d8c94d0854ed19e113c2afa5c49d
       Out:       0
       Signature: 9e0c0839258bff7257ab60fe337616f44c9cb878c92fbceaf169a18ed90608a9af749ecfe3cf4d6ac2e95db8dcb424864bf47c934842ab733568ffeb6424321a
       PubKey:    28f46a4493c79fe6f0a9475de342464b7cc27eae9e259302d925c0f26daf9aa91e07735a8d36ade16c043cd22f1216a91508ecfa38e44a73ec38ba77a394b76e
     Output 0:
       Value:  10
       Script: 3141bb817159aafa64161968f26f162b2bef4b4a


============ Block 00002b40928246808d0b7fe186b63ddb36659e047800b76012d65327be8adad0 ============
Height: 0
Prev. block:
PoW: true

--- Transaction a7b3ddf2cf5658d10d126a73ee74d7b84a04d8c94d0854ed19e113c2afa5c49d:
     Input 0:
       TXID:
       Out:       -1
       Signature:
       PubKey:    43726561746520626c6f636b20636861696e206d616e6e75616c6c79206163636f7264696e6720746f2046756461204d53452050726f6a656374
     Output 0:
       Value:  10
       Script: 4e190c9afd4c7bcb1f09e8263a26ea49e49ced31
```
### 6. 显示账户余额
```
    $> blockchain service -b 15VSra4M24knbrpAfeqSfVEeyY8Qag4GLr
```
*结果*
```
    Balance of '15VSra4M24knbrpAfeqSfVEeyY8Qag4GLr': 40
```