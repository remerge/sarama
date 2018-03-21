#!/bin/sh

set -ex

while ! nc localhost 8474 </dev/null; do echo "Waiting"; sleep 1; done

wget -O/dev/null -S --post-data='{"name":"zk1", "upstream":"localhost:21801", "listen":"0.0.0.0:2181"}' localhost:8474/proxies
#wget -O/dev/null -S --post-data='{"name":"zk2", "upstream":"localhost:21802", "listen":"0.0.0.0:2181"}' localhost:8474/proxies
#wget -O/dev/null -S --post-data='{"name":"zk3", "upstream":"localhost:21803", "listen":"0.0.0.0:2181"}' localhost:8474/proxies
#wget -O/dev/null -S --post-data='{"name":"zk4", "upstream":"localhost:21804", "listen":"0.0.0.0:2181"}' localhost:8474/proxies
#wget -O/dev/null -S --post-data='{"name":"zk5", "upstream":"localhost:21805", "listen":"0.0.0.0:2181"}' localhost:8474/proxies

wget -O/dev/null -S --post-data='{"name":"kafka1", "upstream":"localhost:29091", "listen":"0.0.0.0:9092"}' localhost:8474/proxies
#wget -O/dev/null -S --post-data='{"name":"kafka2", "upstream":"localhost:29092", "listen":"0.0.0.0:9092"}' localhost:8474/proxies
#wget -O/dev/null -S --post-data='{"name":"kafka3", "upstream":"localhost:29093", "listen":"0.0.0.0:9092"}' localhost:8474/proxies
#wget -O/dev/null -S --post-data='{"name":"kafka4", "upstream":"localhost:29094", "listen":"0.0.0.0:9092"}' localhost:8474/proxies
#wget -O/dev/null -S --post-data='{"name":"kafka5", "upstream":"localhost:29095", "listen":"0.0.0.0:9092"}' localhost:8474/proxies
