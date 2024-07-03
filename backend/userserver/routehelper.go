package userserver

type createUserReply struct {
	DakarUserUID string `json:"dakarUserUID"`
}

type msgReply struct {
	Msg string `json:"msg"`
}
