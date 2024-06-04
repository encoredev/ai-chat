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
    AddUserButton
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
import {AddBotModal} from "./AddBotModal";
export const Chat = ({user, channelID}:{user:User, channelID:string}) => {
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
            )]}
        );
        addConversation(conv);
        return conv
    })

    const chatService = service as ExampleChatService;
    useEffect( () => {
        addUser(user)
        setActiveConversation(channelID);
        chatService.joinChannel(channelID);
    },[]);

    useEffect( () => {
        setCurrentUser(user);
    },[user, setCurrentUser]);

    const [userProfile, setUserProfile] = useState<User>();
    const [addUserShow, setAddUserShow] = useState(false);

    const handleChange = (value:string) => {
        // Send typing indicator to the active conversation
        // You can call this method on each onChange event
        // because sendTyping method can throttle sending this event
        // So typing event will not be send to often to the server
        setCurrentMessage(value);
        if ( activeConversation ) {
            sendTyping({
                conversationId: activeConversation?.id,
                isTyping:true,
                userId: user.id,
                content: value, // Note! Most often you don't want to send what the user types, as this can violate his privacy!
                throttle: true
            });
        }
        
    }
    
    const handleSend = (text:string) => {
        
        const message = new ChatMessage({
            id: "", // Id will be generated by storage generator, so here you can pass an empty string
            content: text as unknown as MessageContent<TextContent>,
            contentType: MessageContentType.TextHtml,
            senderId: user.id,
            direction: MessageDirection.Outgoing,
            status: MessageStatus.Sent
        });
        
        if ( activeConversation ) {
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
                                return <TypingIndicator content={`${typingUser.username} is typing`} />
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
                      <ProfileModal show={userProfile !== undefined} user={userProfile}  onHide={() => setUserProfile(undefined)}/>
                      <AddBotModal show={addUserShow} channelID={channelID} onHide={() => setAddUserShow(false)}/>
                      <Sidebar position="left" scrollable>
                          {activeConversation?.participants.map((p) =>
                            <ConversationHeader style={{backgroundColor: "#fff"}} onClick={() => setUserProfile(getUser(p.id))}>
                                <Avatar src={getUser(p.id)?.avatar}/>
                                <ConversationHeader.Content>
                                    {getUser(p.id)?.username}
                                </ConversationHeader.Content>
                            </ConversationHeader>
                          )}
                          <AddUserButton onClick={()=>setAddUserShow(true)}>
                              Add Bot
                          </AddUserButton>
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
