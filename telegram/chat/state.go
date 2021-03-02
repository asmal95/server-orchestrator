package chat

var states = make(map[int64]State)

type State struct {
	CurrentControlMessage int64  // Message ID with active button. Before changing of the control need to remove buttons from old control-message.
	Action                string // Current action: control something, edit something (edit or none?)
	Meta                  map[string]string
}

func GetState(chatId int64) State {
	if state, ok := states[chatId]; ok {
		return state
	} else {
		state := State{
			CurrentControlMessage: -1,
			Action:                "none",
			Meta:                  make(map[string]string),
		}
		states[chatId] = state
		return state
	}
}

func SetControlMessage(chatId int64, messageId int64) State {
	state := GetState(chatId)
	state.CurrentControlMessage = messageId
	states[chatId] = state
	return state
}

func SetAction(chatId int64, action string) State {
	state := GetState(chatId)
	state.Action = action
	states[chatId] = state
	return state
}

func SetMeta(chatId int64, key string, value string) State {
	state := GetState(chatId)
	state.Meta[key] = value
	states[chatId] = state
	return state
}
