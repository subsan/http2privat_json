package connector

func CheckConnection() JsonEntity {
	if !isConnected {
		return JsonEntity{
			Error:            true,
			ErrorDescription: "Terminal not connected",
		}
	}

	return SyncSender(JsonEntity{
		Method: "PingDevice",
	})
}
