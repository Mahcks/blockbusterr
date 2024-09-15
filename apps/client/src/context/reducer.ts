import {
    REDUCER_ACTION_TYPE,
    ReducerAction,
    StateType,
} from '@/context/WebsocketContext';

export const reducer = (state: StateType, action: ReducerAction): StateType => {
    if (action.type === REDUCER_ACTION_TYPE.WS_CONNECTED) {
        return {
            ...state,
            wsConnected: true,
        };
    } else if (action.type === REDUCER_ACTION_TYPE.WS_DISCONNECTED) {
        return {
            ...state,
            wsConnected: false,
        };
    }

    throw new Error();
};