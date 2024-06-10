import { FC, useCallback, useEffect, useState } from "react";

import {
  Avatar,
  AvatarGroup,
  ChatContainer,
  ConversationHeader,
  MainContainer,
  Message,
  MessageGroup,
  MessageInput,
  MessageList,
  Sidebar,
  TypingIndicator,
} from "@chatscope/chat-ui-kit-react";

import {
  ChatMessage,
  Conversation as Conv,
  MessageContent,
  MessageContentType,
  MessageDirection,
  MessageStatus,
  Participant,
  TextContent,
  useChat,
  User,
} from "@chatscope/use-chat";
import { ExampleChatService } from "./ChatService";
import ProfileModal from "./ProfileModal";
import AddBotModal, { AddBotStatus } from "./AddBotModal";
import {
  DiscordLogo,
  List,
  Robot,
  SlackLogo,
  Spinner,
  User as UserIcon,
  WarningCircle,
} from "@phosphor-icons/react";
import poweredBy from "../assets/powered-by-encore.png";
import InviteFriendModal from "./InviteFriendModal.tsx";
import SlideOver from "./SlideOver.tsx";

const WHITE = "#dcdee1";

export const Chat = ({
  user,
  channelID,
}: {
  user: User;
  channelID: string;
}) => {
  // Get all chat related values and methods from useChat hook
  const {
    addConversation,
    currentMessages,
    activeConversation,
    setActiveConversation,
    sendMessage,
    getUser,
    currentMessage,
    setCurrentMessage,
    sendTyping,
    setCurrentUser,
    service,
    addUser,
  } = useChat();

  useState(() => {
    let conv = new Conv({
      id: channelID,
      participants: [new Participant({ id: user.id })],
    });
    addConversation(conv);
    return conv;
  });

  const [botStatus, setBotStatus] = useState<AddBotStatus | undefined>();

  const chatService = service as ExampleChatService;
  useEffect(() => {
    addUser(user);
    setActiveConversation(channelID);
    chatService.joinChannel(channelID);
  }, []);

  useEffect(() => {
    setCurrentUser(user);
  }, [user, setCurrentUser]);

  const [userProfile, setUserProfile] = useState<User>();
  const [addUserShow, setAddUserShow] = useState(false);
  const [inviteFriendShow, setInviteFriend] = useState(false);
  const [mobileSidebarShow, setMobileSidebarShow] = useState(false);

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
        throttle: true,
      });
    }
  };

  const handleSend = (text: string) => {
    const message = new ChatMessage({
      id: "", // Id will be generated by storage generator, so here you can pass an empty string
      content: text as unknown as MessageContent<TextContent>,
      contentType: MessageContentType.TextHtml,
      senderId: user.id,
      direction: MessageDirection.Outgoing,
      status: MessageStatus.Sent,
    });

    if (activeConversation) {
      sendMessage({
        message,
        conversationId: activeConversation.id,
        senderId: user.id,
      });
    }
  };

  const getTypingIndicator = useCallback(() => {
    if (activeConversation) {
      const typingUsers = activeConversation.typingUsers;

      if (typingUsers.length > 0) {
        const typingUserId = typingUsers.items[0].userId;

        // Check if typing user participates in the conversation
        if (activeConversation.participantExists(typingUserId)) {
          const typingUser = getUser(typingUserId);

          if (typingUser) {
            return (
              <TypingIndicator
                className="!bg-transparent !px-4"
                content={`${typingUser.username} is typing`}
              />
            );
          }
        }
      }
    }

    return null;
  }, [activeConversation, getUser]);

  const getSidebar = () => {
    return (
      <ChatSidebar
        user={user}
        getUser={getUser}
        setUserProfile={(user) => {
          setUserProfile(user);
          setMobileSidebarShow(false);
        }}
        botStatus={botStatus}
        activeConversation={activeConversation}
        showAddBotModal={() => {
          setAddUserShow(true);
          setMobileSidebarShow(false);
        }}
        showInviteFriendModal={() => {
          setInviteFriend(true);
          setMobileSidebarShow(false);
        }}
      />
    );
  };

  return (
    <MainContainer responsive className="w-full h-screen !border-none">
      <ProfileModal
        show={userProfile !== undefined}
        user={userProfile}
        onHide={() => setUserProfile(undefined)}
      />

      <AddBotModal
        statusChange={(s?: AddBotStatus) => setBotStatus(s)}
        show={addUserShow}
        channelID={channelID}
        onHide={() => setAddUserShow(false)}
      />

      <InviteFriendModal
        show={inviteFriendShow}
        channelID={channelID}
        onHide={() => setInviteFriend(false)}
      />

      {/* Desktop sidebar */}
      {getSidebar()}

      {/* Mobile sidebar */}
      <SlideOver
        show={mobileSidebarShow}
        onHide={() => setMobileSidebarShow(false)}
      >
        {getSidebar()}
      </SlideOver>

      <ChatContainer>
        <ConversationHeader className="!bg-gray-900 !p-0 !border-gray-500 shadow-lg">
          <ConversationHeader.Back className="ml-4">
            {botStatus?.status === "creating" ? (
              <Spinner
                size={30}
                color={WHITE}
                className="animate-spin"
                onClick={() => setMobileSidebarShow(true)}
              />
            ) : (
              <List
                size={30}
                color={WHITE}
                onClick={() => setMobileSidebarShow(true)}
              />
            )}
          </ConversationHeader.Back>
          <ConversationHeader.Content>
            <div className="flex items-center justify-between mr-4">
              <p className="flex font-mono items-center space-x-2 text-white font-semibold px-4 py-2">
                <span className="opacity-50 text-xl">#</span>{" "}
                <span className="text-sm">{channelID}</span>
              </p>

              <ConversationHeader.Back className="ml-4">
                <AvatarGroup size="sm">
                  {activeConversation?.participants.map((p) => {
                    const isBot = !!getUser(p.id)?.avatar;
                    return (
                      <Avatar src={getUser(p.id)?.avatar}>
                        {!isBot && <ProfileCircle user={user} size="sm" />}
                      </Avatar>
                    );
                  })}
                </AvatarGroup>
              </ConversationHeader.Back>
            </div>
          </ConversationHeader.Content>
        </ConversationHeader>
        <MessageList
          typingIndicator={getTypingIndicator()}
          className="!bg-gray-800"
        >
          {currentMessages.map((g) => (
            <MessageGroup key={g.id}>
              <MessageGroup.Messages>
                {g.messages.map((m: ChatMessage<MessageContentType>) => {
                  const isBot = !!getUser(m.senderId)?.avatar;
                  return (
                    <Message
                      key={m.id}
                      className="w-full !p-2 hover:bg-gray-700 transition"
                      model={{
                        type: "custom",
                        direction: "incoming",
                        position: "normal",
                      }}
                      avatarPosition="tl"
                    >
                      <Avatar
                        name={m.senderId}
                        status="available"
                        src={getUser(m.senderId)?.avatar}
                      >
                        {!isBot && <ProfileCircle user={user} size="md" />}
                      </Avatar>
                      <Message.CustomContent className="-mt-2 text-white">
                        <p className="flex items-center space-x-3">
                          <span className="font-semibold">
                            {getUser(m.senderId)?.username}
                          </span>
                          <span className="text-gray-500 text-xs">
                            {getTodaysDate()}
                          </span>
                        </p>
                        <p>{m.content as unknown as string}</p>
                      </Message.CustomContent>
                    </Message>
                  );
                })}
              </MessageGroup.Messages>
            </MessageGroup>
          ))}
        </MessageList>
        <MessageInput
          className="!bg-gray-800 !border-none !pb-4 !pr-3"
          value={currentMessage}
          onChange={handleChange}
          onSend={handleSend}
          disabled={!activeConversation}
          attachButton={false}
          placeholder="Type here..."
        />
      </ChatContainer>
    </MainContainer>
  );
};

const ChatSidebar: FC<{
  activeConversation?: Conv<any> | undefined;
  user: User;
  setUserProfile: (user?: User) => void;
  getUser: (userId: string) => User | undefined;
  botStatus: AddBotStatus | undefined;
  showAddBotModal: () => void;
  showInviteFriendModal: () => void;
}> = ({
  activeConversation,
  user,
  setUserProfile,
  getUser,
  botStatus,
  showAddBotModal,
  showInviteFriendModal,
}) => {
  return (
    <Sidebar
      position="left"
      scrollable
      className="flex justify-between !bg-gray-900 !border-none !max-w-[220px] !shadow-xl py-2"
      style={{
        flex: "35%",
      }}
    >
      <div>
        {activeConversation?.participants.map((p) => {
          const isBot = !!getUser(p.id)?.avatar;
          return (
            <div
              id={p.id}
              className={`
                  flex space-x-2 items-center px-4 py-3 hover:!bg-gray-800
                  ${isBot ? "cursor-pointer" : "cursor-default"}
                `}
              onClick={() => {
                if (isBot) setUserProfile(getUser(p.id));
              }}
            >
              <Avatar status="available" src={getUser(p.id)?.avatar}>
                {!isBot && <ProfileCircle user={user} size="md" />}
              </Avatar>
              <div>
                <span className="text-white font-semibold">
                  {getUser(p.id)?.username}
                </span>
              </div>
            </div>
          );
        })}

        {botStatus?.status === "creating" && (
          <div className="flex space-x-2 items-center p-4 py-3 cursor-pointer">
            <Avatar status="eager" size="md">
              <Spinner size={40} color={WHITE} className="animate-spin" />
            </Avatar>
            <div className="text-white">Creating {botStatus?.botName}</div>
          </div>
        )}

        {botStatus?.status === "failure" && (
          <div className="flex space-x-2 items-center p-4 py-3 cursor-pointer">
            <Avatar status="dnd" size="md">
              <WarningCircle size={40} color={WHITE} />
            </Avatar>
            <div className="text-white">
              Failed to create {botStatus?.botName}
            </div>
          </div>
        )}

        <div className="h-[2px] w-5/6 bg-gray-700 mx-auto my-4" />

        <div
          onClick={showAddBotModal}
          className="flex items-center space-x-2 text-white px-4 py-2 cursor-pointer hover:!bg-gray-800"
        >
          <div>
            <Robot size={25} color={WHITE} />
          </div>
          <span className="uppercase text-xs">Add Bot</span>
        </div>

        <div
          onClick={showInviteFriendModal}
          className="flex items-center space-x-2 text-white px-4 py-2 cursor-pointer hover:!bg-gray-800"
        >
          <div>
            <UserIcon size={25} color={WHITE} />
          </div>
          <span className="uppercase text-xs">Invite Friend</span>
        </div>
      </div>

      <div className="px-4 pb-2">
        <a
          href="https://github.com/encoredev/ai-chat/tree/main?tab=readme-ov-file#adding-a-discord-bot"
          target="_blank"
        >
          <div className="flex items-center text-white text-sm space-x-2 mb-4 opacity-70 cursor-pointer hover:opacity-100">
            <DiscordLogo className="w-6 h-6" />
            <p>Add to your Discord</p>
          </div>
        </a>

        <a
          href="https://github.com/encoredev/ai-chat/tree/main?tab=readme-ov-file#adding-a-slack-bot"
          target="_blank"
        >
          <div className="flex items-center text-white text-sm space-x-2 mb-4 opacity-70 cursor-pointer hover:opacity-100">
            <SlackLogo className="w-6 h-6" />
            <p>Add to your Slack</p>
          </div>
        </a>

        <a
          href="https://github.com/encoredev/ai-chat/tree/main"
          target="_blank"
        >
          <img
            src={poweredBy}
            alt="Powered by Encore"
            className="block mx-auto"
          />
        </a>
      </div>
    </Sidebar>
  );
};

const ProfileCircle: FC<{ user?: User; size: "md" | "sm" }> = ({
  user,
  size,
}) => {
  return (
    <figure
      className={`
        flex items-center justify-center rounded-full bg-blue relative uppercase text-white text-lg font-semibold
        ${size === "md" ? "h-[40px] w-[40px]" : "h-[26px] w-[26px]"}
      `}
    >
      {size === "md" && (
        <>
          {user?.username[0]}
          {user?.username[1]}
        </>
      )}
    </figure>
  );
};

const getTodaysDate = () => {
  const date = new Date();
  const [year, month, day] = date.toISOString().split("T")[0].split("-");
  return `${day}/${month}/${year}`;
};