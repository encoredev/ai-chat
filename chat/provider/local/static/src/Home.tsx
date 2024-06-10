import {
  createSearchParams,
  useNavigate,
  useSearchParams,
} from "react-router-dom";
import { useState } from "react";
import { humanId } from "human-id";
import poweredBy from "./assets/powered-by-encore.png";
import Button from "./components/Button";

export const Home = () => {
  let [username, setUsername] = useState("");
  const [status, setStatus] = useState("typing");
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const channelID = searchParams.get("channel");

  async function joinChat() {
    setStatus("submitting");
    navigate({
      pathname: "/chat",
      search: createSearchParams({
        name: username,
        channel: channelID || humanId(),
      }).toString(),
    });
  }

  return (
    <div className="px-6 py-24 sm:py-32 lg:px-8">
      <div className="mx-auto max-w-2xl text-center">
        <h2 className="text-4xl font-bold tracking-tight text-white sm:text-6xl">
          AI Chat
        </h2>
        <p className="mt-6 text-lg leading-8 text-gray-300">
          Chat with AI generated bots and create new bots.
        </p>
      </div>

      <form
        className="mx-auto flex items-center justify-center w-full mt-10 space-x-3"
        onSubmit={(event) => {
          event.preventDefault();
          joinChat();
        }}
      >
        <input
          type="text"
          placeholder="Your name"
          className="max-w-72 text-xl block w-full rounded-md border-gray-500 bg-gray-800 focus:ring-0 focus:border-gray-500"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
        />

        <Button
          mode="light"
          size="lg"
          type="submit"
          disabled={!username || status === "submitting"}
        >
          Join {channelID ? "channel " + channelID : "Chat"}
        </Button>
      </form>

      <div className="absolute bottom-3 left-0 max-w-52">
        <a href="https://github.com/encoredev/ai-chat/tree/main">
          <img
            src={poweredBy}
            alt="Powered by Encore"
            className="block mx-auto"
          />
        </a>
      </div>
    </div>
  );
};
