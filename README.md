# distccgo
the purpose of this project is to make a distriuted tool using go language to accelerate the compile speed like distcc
usercmd:
make all -j16 CXX=distccgo gcc

developer command
sudo cp distccgo /bin/distccgo
dameon useage：
./distcc dameon -s 127.0.0.1 -c -l 127.0.0.1

 go build -a

