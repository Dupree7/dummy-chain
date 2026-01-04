# Some calls can be made using curl

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc": "2.0", "method": "chain.GetAccountInfo", "params": ["address"], "id": 1}' localhost:12345

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc": "2.0", "method": "chain.GetTransactionByHash", "params": ["hash"], "id": 1}' localhost:12345

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc": "2.0", "method": "chain.GetCurrenBlockHeight", "params": [1], "id": 1}' localhost:12345

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc": "2.0", "method": "chain.GetBlockByHash", "params": ["hash"], "id": 1}' localhost:12345

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc": "2.0", "method": "chain.GetBlockByHeight", "params": [2], "id": 1}' localhost:12345

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc": "2.0", "method": "chain.GetBlocksInterval", "params": [{"Left": 1, "Right": 3}], "id": 1}' localhost:12345

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc": "2.0", "method": "chain.SendTransaction", "params": ["tx"], "id": 1}' localhost:12345