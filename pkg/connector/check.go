package connector

func CheckConnection() JsonEntity {
    return SyncSender(JsonEntity{
        Method: "PingDevice",
    })
}
