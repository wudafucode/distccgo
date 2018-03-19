# distccgo
the purpose of this project is to make a distriuted tool using go language to accelerate the compile speed like distcc

compile cmd:
make all -j16 CXX=distccgo gcc

developer command

1  sudo cp distccgo /bin/distccgo
2  ./distcc server  -h 192.168.16.214 ./data
3  ./distcc worker -masternode 192.168.16.214 -h 192.168.16.211
4  ./distcc monitor----this command has to be exectued on every machine you want to compile
 
compile cmd 
go build -a
http://127.0.0.1:4001/cluster/status
