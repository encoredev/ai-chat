// This IChatService implementation is only an example and has no real business value.
// However, this is good start point to make your own implementation.
// Using this service it's possible to connects two or more chats in the same application for a demonstration purposes

import {
  Conversation,
  ConversationId,
  ConversationRole,
  IChatService, MessageStatus,
  Participant,
  TypingUsersList
} from "@chatscope/use-chat";
import { ChatEventType, MessageContentType, MessageDirection } from "@chatscope/use-chat/dist/enums";
import ReconnectingWebSocket from "reconnecting-websocket";
import {
  ChatEventHandler,
  SendMessageServiceParams,
  SendTypingServiceParams,
  UpdateState,
} from "@chatscope/use-chat/dist/Types";
import { IStorage } from "@chatscope/use-chat/dist/interfaces";
import { ChatEvent, MessageEvent, UserTypingEvent } from "@chatscope/use-chat/dist/events";
import { ChatMessage } from "@chatscope/use-chat/dist/ChatMessage";

type EventHandlers = {
  onMessage: ChatEventHandler<
    ChatEventType.Message,
    ChatEvent<ChatEventType.Message>
  >;
  onConnectionStateChanged: ChatEventHandler<
    ChatEventType.ConnectionStateChanged,
    ChatEvent<ChatEventType.ConnectionStateChanged>
  >;
  onUserConnected: ChatEventHandler<
    ChatEventType.UserConnected,
    ChatEvent<ChatEventType.UserConnected>
  >;
  onUserDisconnected: ChatEventHandler<
    ChatEventType.UserDisconnected,
    ChatEvent<ChatEventType.UserDisconnected>
  >;
  onUserPresenceChanged: ChatEventHandler<
    ChatEventType.UserPresenceChanged,
    ChatEvent<ChatEventType.UserPresenceChanged>
  >;
  onUserTyping: ChatEventHandler<
    ChatEventType.UserTyping,
    ChatEvent<ChatEventType.UserTyping>
  >;
  [key: string]: any;
};

interface ServiceMessage {
  type: string;
  userId: string;
  conversationId: ConversationId;
  content: string;
  isTyping: boolean;
}

export class ExampleChatService implements IChatService {
  storage?: IStorage;
  updateState: UpdateState;
  ws: ReconnectingWebSocket;

  eventHandlers: EventHandlers = {
    onMessage: () => {},
    onConnectionStateChanged: () => {},
    onUserConnected: () => {},
    onUserDisconnected: () => {},
    onUserPresenceChanged: () => {},
    onUserTyping: () => {},
  };

  constructor(storage: IStorage, update: UpdateState, conversationID:string, userID:string) {
    this.storage = storage;
    this.updateState = update;
    var proto = document.location.protocol === "https:" ? "wss://" : "ws://";
    this.ws = new ReconnectingWebSocket(`ws://localhost:4000/encorechat/channels/${conversationID}/subscribe/${userID}`);

    this.ws.addEventListener("message", (event) => {
      let msg: ServiceMessage = JSON.parse(event.data)
      if (msg.type === "join") {

      } else if (msg.type === "message") {
        if (msg.userId === userID) {
          return;
        }
        let message = new ChatMessage<MessageContentType.TextPlain>({
          id: "",
          content: msg,
          contentType: MessageContentType.TextPlain,
          senderId: msg.userId,
          direction: MessageDirection.Incoming,
          status: MessageStatus.Pending
        });
        const conversationId = msg.conversationId;
        if (this.eventHandlers.onMessage) {
          if (this.storage === undefined) {
            return;
          }
          let [c, _] = this.storage.getConversation(conversationId);
          if (c === undefined) {
            this.storage?.addConversation(new Conversation({
              id: conversationId,
              participants: [
                new Participant({
                  id: msg.userId,
                  role: new ConversationRole([])
                })
              ],
              unreadCounter: 0,
              typingUsers: new TypingUsersList({items: []}),
              draft: ""
            }));
          }

          this.eventHandlers.onMessage(
            new MessageEvent({ message, conversationId })
          );
        }
      } else if ( msg.type === "typing") {
        const { userId, isTyping, conversationId, content } = msg;

        if (this.eventHandlers.onUserTyping) {
          // Running the onUserTyping callback registered by ChatProvider will cause:
          // 1. Add the user to the list of users who are typing in the conversation
          // 2. Debounce
          // 3. Re-render
          this.eventHandlers.onUserTyping(
            new UserTypingEvent({
              userId,
              isTyping,
              conversationId,
              content,
            })
          );
        }
      }
    });
  }

  sendMessage({ message, conversationId }: SendMessageServiceParams) {
    if (message.contentType == MessageContentType.TextHtml) {
      const msg = message as ChatMessage<MessageContentType.TextHtml>;
      this.ws.send(msg.content?.toString());
    }

  }

  sendTyping({
               isTyping,
               content,
               conversationId,
               userId,
             }: SendTypingServiceParams) {
  }

  // The ChatProvider registers callbacks with the service.
  // These callbacks are necessary to notify the provider of the changes.
  // For example, when your service receives a message, you need to run an onMessage callback,
  // because the provider must know that the new message arrived.
  // Here you need to implement callback registration in your service.
  // You can do it in any way you like. It's important that you will have access to it elsewhere in the service.
  on<T extends ChatEventType, H extends ChatEvent<T>>(
    evtType: T,
    evtHandler: ChatEventHandler<T, H>
  ) {
    const key = `on${evtType.charAt(0).toUpperCase()}${evtType.substring(1)}`;

    if (key in this.eventHandlers) {
      this.eventHandlers[key] = evtHandler;
    }
  }

  // The ChatProvider can unregister the callback.
  // In this case remove it from your service to keep it clean.
  off<T extends ChatEventType, H extends ChatEvent<T>>(
    evtType: T,
    eventHandler: ChatEventHandler<T, H>
  ) {
    const key = `on${evtType.charAt(0).toUpperCase()}${evtType.substring(1)}`;
    if (key in this.eventHandlers) {
      this.eventHandlers[key] = () => {};
    }
  }
}
