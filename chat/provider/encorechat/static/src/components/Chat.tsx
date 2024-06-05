import {useMemo, useCallback, useEffect, useState} from "react";

import {
  MainContainer,
  Sidebar,
  Avatar,
  ChatContainer,
  ConversationHeader,
  MessageGroup,
  Message,
  MessageList,
  MessageInput,
  TypingIndicator,
  AddUserButton, Loader
} from "@chatscope/chat-ui-kit-react";

import {
  useChat,
  ChatMessage,
  MessageContentType,
  MessageDirection,
  MessageStatus, Participant,
} from "@chatscope/use-chat";
import {MessageContent, TextContent, User, Conversation as Conv} from "@chatscope/use-chat";
import {ExampleChatService} from "./ChatService";
import {ProfileModal} from "./ProfileModal";
import {AddBotModal, AddBotStatus} from "./AddBotModal";

export const Chat = ({user, channelID}: { user: User, channelID: string }) => {
  // Get all chat related values and methods from useChat hook
  const {
    addConversation,
    currentMessages, conversations,
    activeConversation,
    setActiveConversation,
    sendMessage,
    getUser,
    currentMessage,
    setCurrentMessage,
    sendTyping,
    setCurrentUser,
    service,
    addUser
  } = useChat();

  useState(() => {
    let conv = new Conv({
        id: channelID,
        participants: [new Participant(
          {id: user.id}
        )]
      }
    );
    addConversation(conv);
    return conv
  })

  const [botStatus, setBotStatus] = useState<AddBotStatus>();

  const chatService = service as ExampleChatService;
  useEffect(() => {
    addUser(user)
    setActiveConversation(channelID);
    chatService.joinChannel(channelID);
  }, []);

  useEffect(() => {
    setCurrentUser(user);
  }, [user, setCurrentUser]);

  const [userProfile, setUserProfile] = useState<User>();
  const [addUserShow, setAddUserShow] = useState(false);

  const handleChange = (value: string) => {
    // Send typing indicator to the active conversation
    // You can call this method on each onChange event
    // because sendTyping method can throttle sending this event
    // So typing event will not be send to often to the server
    setCurrentMessage(value);
    if (activeConversation) {
      sendTyping({
        conversationId: activeConversation?.id,
        isTyping: true,
        userId: user.id,
        content: value, // Note! Most often you don't want to send what the user types, as this can violate his privacy!
        throttle: true
      });
    }

  }

  const handleSend = (text: string) => {

    const message = new ChatMessage({
      id: "", // Id will be generated by storage generator, so here you can pass an empty string
      content: text as unknown as MessageContent<TextContent>,
      contentType: MessageContentType.TextHtml,
      senderId: user.id,
      direction: MessageDirection.Outgoing,
      status: MessageStatus.Sent
    });

    if (activeConversation) {
      sendMessage({
        message,
        conversationId: activeConversation.id,
        senderId: user.id,
      });
    }

  };

  const getTypingIndicator = useCallback(
    () => {

      if (activeConversation) {

        const typingUsers = activeConversation.typingUsers;

        if (typingUsers.length > 0) {

          const typingUserId = typingUsers.items[0].userId;

          // Check if typing user participates in the conversation
          if (activeConversation.participantExists(typingUserId)) {

            const typingUser = getUser(typingUserId);

            if (typingUser) {
              return <TypingIndicator content={`${typingUser.username} is typing`}/>
            }

          }

        }

      }


      return undefined;

    }, [activeConversation, getUser],
  );

  return (

    <MainContainer responsive
                   style={{
                     height: '600px'
                   }}>
      <ProfileModal show={userProfile !== undefined} user={userProfile} onHide={() => setUserProfile(undefined)}/>
      <AddBotModal statusChange={(s:AddBotStatus) => setBotStatus(s)} show={addUserShow}
                   channelID={channelID} onHide={() => setAddUserShow(false)}/>
      <Sidebar position="left" scrollable>
        {activeConversation?.participants.map((p) =>
          <ConversationHeader style={{backgroundColor: "#fff"}} onClick={() => setUserProfile(getUser(p.id))}>
            <Avatar src={getUser(p.id)?.avatar}/>
            <ConversationHeader.Content>
              {getUser(p.id)?.username}
            </ConversationHeader.Content>
          </ConversationHeader>
        )}
        {botStatus?.status === "creating" ? (
            <ConversationHeader style={{backgroundColor: "#fff"}}>
              <Avatar>
                <Loader/>
              </Avatar>
              <ConversationHeader.Content>
                Creating {botStatus.botName}
              </ConversationHeader.Content>
            </ConversationHeader>
        ) : botStatus?.status === "failure" ? (
          <ConversationHeader style={{backgroundColor: "#fff"}}>
            <Avatar>
              <svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" fill="#FF1111"
                   className="bi bi-x-octagon" viewBox="0 0 16 16">
                <path
                  d="M4.54.146A.5.5 0 0 1 4.893 0h6.214a.5.5 0 0 1 .353.146l4.394 4.394a.5.5 0 0 1 .146.353v6.214a.5.5 0 0 1-.146.353l-4.394 4.394a.5.5 0 0 1-.353.146H4.893a.5.5 0 0 1-.353-.146L.146 11.46A.5.5 0 0 1 0 11.107V4.893a.5.5 0 0 1 .146-.353zM5.1 1 1 5.1v5.8L5.1 15h5.8l4.1-4.1V5.1L10.9 1z"/>
                <path
                  d="M4.646 4.646a.5.5 0 0 1 .708 0L8 7.293l2.646-2.647a.5.5 0 0 1 .708.708L8.707 8l2.647 2.646a.5.5 0 0 1-.708.708L8 8.707l-2.646 2.647a.5.5 0 0 1-.708-.708L7.293 8 4.646 5.354a.5.5 0 0 1 0-.708"/>
              </svg>
            </Avatar>
            <ConversationHeader.Content>
              Failed to create {botStatus.botName}
            </ConversationHeader.Content>
          </ConversationHeader>
        ) : (
          <AddUserButton onClick={() => setAddUserShow(true)}>
            Add Bot
          </AddUserButton>
        )}

      </Sidebar>

      <ChatContainer>
        <MessageList typingIndicator={getTypingIndicator()}>
          {activeConversation && currentMessages.map((g) => <MessageGroup key={g.id}
                                                                          direction={g.direction}>
            <MessageGroup.Messages>
            {g.messages.map((m: ChatMessage<MessageContentType>) =>
                <Message key={m.id} model={{
                  type: "html",
                  payload: m.content,
                  direction: m.direction,
                  position: "normal"
                }}>
                  {m.direction === MessageDirection.Incoming && m.senderId !== user.id &&
                    <Avatar src={getUser(m.senderId)?.avatar} name={m.senderId}/>}
                </Message>
              )}
            </MessageGroup.Messages>
          </MessageGroup>)}
        </MessageList>
        <MessageInput value={currentMessage} onChange={handleChange} onSend={handleSend}
                      disabled={!activeConversation} attachButton={false} placeholder="Type here..."/>
      </ChatContainer>

    </MainContainer>);
}
