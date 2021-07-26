package socket

import "errors"

var Err_Conncet_Unknown = errors.New("unknown")
var Err_Max_Buff_Size = errors.New("buff size must less 4M")
var Err_Socket_Not_Open = errors.New("socket not open")
var Err_Conn_Is_Nil = errors.New("conn is nil")
var Err_Handler_Packet_Fn = errors.New("handler packet func must not nil")
var Err_Send_Packet_Is_Nil = errors.New("send ni packet")
