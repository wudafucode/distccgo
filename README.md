# distccgo
the purpose of this project is to make a distriuted tool using go language to accelerate the compile speed like distcc
usercmd:
make all -j16 CXX=distccgo gcc

developer command
sudo cp distccgo /bin/distccgo
dameon useage：
./distcc worker -s 127.0.0.1 -c -l 127.0.0.1
./distcc server ./data
./distcc worker 
 go build -a
http://127.0.0.1:4001/cluster/status
