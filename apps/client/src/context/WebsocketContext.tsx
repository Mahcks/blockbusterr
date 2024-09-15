/* eslint-disable @typescript-eslint/no-explicit-any */
import React, { ReactElement } from "react";

import { reducer } from "@/context/reducer";
import { useWebSocket } from "react-use-websocket/dist/lib/use-websocket";
import {
  Code,
  HeartbeatPayload,
  HelloPayload,
  Message,
  NewMessage,
} from "@/types/message";

export type StateType = {
  loaded: boolean;
  error: null | string;
  wsConnected: boolean;
};

const initState: StateType = {
  loaded: false,
  error: null,
  wsConnected: false,
};

export const enum REDUCER_ACTION_TYPE {
  WS_CONNECTED,
  WS_DISCONNECTED,
}

export type ReducerAction = {
  type: REDUCER_ACTION_TYPE;
  payload?: any;
};

const useWebsocketContext = (initState: StateType) => {
  const [state, dispatch] = React.useReducer(reducer, initState);

  const { sendJsonMessage } = useWebSocket(
    `${import.meta.env.VITE_API_URL}/ws`,
    {
      onOpen: () => {
        dispatch({
          type: REDUCER_ACTION_TYPE.WS_CONNECTED,
        });
      },
      onMessage: (event) => {
        const e: Message = JSON.parse(event.data);

        switch (e.c) {
          case Code.CodeHello: {
            const helloPayload: HelloPayload = e.d;
            console.debug(
              `[WS] Connected with session ID ${helloPayload.session_id}`
            );
            // TODO: Send sendJsonMessage with Code.CodeSubscribe
            break;
          }

          case Code.CodeHeartbeat: {
            const heartbeatPayload: HeartbeatPayload = e.d;
            console.debug(`[WS] Heartbeat ${heartbeatPayload.count} received`);

            sendJsonMessage(
              NewMessage(Code.CodeHeartbeat, {
                count: heartbeatPayload.count + 1,
              })
            );

            console.debug(`[WS] Heartbeat ${heartbeatPayload.count} sent`);
            break;
          }

          default:
            break;
        }
      },
      onError: (event) => console.log(event),
      onClose: (event) => {
        console.log(event);
        dispatch({
          type: REDUCER_ACTION_TYPE.WS_DISCONNECTED,
        });
      },
      shouldReconnect: () => true,
    }
  );

  return { state, dispatch };
};

type UseWebSocketContextType = ReturnType<typeof useWebsocketContext>;

const initContextState: UseWebSocketContextType = {
  state: initState,
  dispatch: () => {
    throw new Error("Function not implemented yet");
  },
};

export const WebSocketContext =
  React.createContext<UseWebSocketContextType>(initContextState);

type ChildrenType = {
  children?: ReactElement | ReactElement[] | undefined;
};

export const WebsocketProvider = ({ children }: ChildrenType): ReactElement => {
  return (
    <WebSocketContext.Provider value={useWebsocketContext(initState)}>
      {children}
    </WebSocketContext.Provider>
  );
};
