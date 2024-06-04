import {useMemo, useCallback, useEffect} from "react";

import { MainContainer, Sidebar, ConversationList, Conversation, Avatar, ChatContainer, ConversationHeader, MessageGroup, Message,MessageList, MessageInput, TypingIndicator } from "@chatscope/chat-ui-kit-react";

import anonAvatar from "./assets/anon.png";

import {
  useChat,
  ChatMessage,
  MessageContentType,
  MessageDirection,
  Participant,
  ConversationRole,
  Presence, UserStatus,
  MessageStatus, ConversationId, ChatProvider, IStorage, UpdateState, BasicStorage, TypingUsersList
} from "@chatscope/use-chat";
import {MessageContent, TextContent, User, Conversation as Conv} from "@chatscope/use-chat";
import {Container, Row, Col} from "react-bootstrap";
import {ExampleChatService} from "./components/ChatService";
import {Chat} from "./components/Chat";
import {AutoDraft} from "@chatscope/use-chat/dist/enums/AutoDraft";
import {nanoid} from "nanoid";

export const EncoreChat = ({userName, channelID}:{userName:string, channelID:string}) => {
  const messageIdGenerator = (message: ChatMessage<MessageContentType>) => nanoid();
  const groupIdGenerator = () => nanoid();
  const user = new User({
    id: userName,
    presence: new Presence({status: UserStatus.Available, description: ""}),
    firstName: "",
    lastName: "",
    username: userName,
    email: "",
    avatar: anonAvatar,
    bio: ""
  });
  const serviceFactory = (storage: IStorage, updateState: UpdateState) => {
    return new ExampleChatService(storage, updateState, user);
  };
  const storage = new BasicStorage({groupIdGenerator, messageIdGenerator})
  return (
    <div className="d-flex flex-column overflow-hidden">
      <Container fluid className="p-4 flex-grow-1 position-relative overflow-hidden">
            <ChatProvider serviceFactory={serviceFactory} storage={storage} config={{
              typingThrottleTime: 250,
              typingDebounceTime: 5000,
              debounceTyping: true,
              autoDraft: AutoDraft.Save | AutoDraft.Restore
            }}>
              <Chat channelID={channelID} user={user}/>
            </ChatProvider>
      </Container>
    </div>);
}