package coswss

var Transform = struct {
	Encode func(b []byte) ([]byte, error) //将WS的二进制转换成TCP二进制
	Decode func(b []byte) ([]byte, error) //将TCP的二进制转换成WS二进制
}{}
