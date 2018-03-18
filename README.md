# distccgo
the purpose of this project is to make a distriuted tool using go language to accelerate the compile speed like distcc
usercmd:
make all -j16 CXX=distccgo gcc

developer command
sudo cp distccgo /bin/distccgo
dameon useage：

./distcc server ./data
./distcc worker 

./distcc monitor----this command has to be exectued on every machine you want to compile
 
compile cmd 
go build -a
http://127.0.0.1:4001/cluster/status
