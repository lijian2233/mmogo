package protocal

const (
	op_ping         = 20
	op_pong         = 21
	op_gate2login   = 100 //gateway to login
	op_login2gate   = 101 //login to gateway
	op_login2center = 102 //login to center
	op_center2login = 103 //center to
	op_gate2game    = 104
	op_game2gate    = 105
	op_login2db     = 106
	op_db2login     = 107
	op_center2db    = 108
	op_db2center    = 109
	op_game2db      = 110
	op_db2game      = 111
)
