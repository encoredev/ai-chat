import { FC, useState } from "react";
import { DialogTitle } from "@headlessui/react";
import Modal, { ModalProps } from "./Modal.tsx";
import Button from "./Button.tsx";

export interface AddBotStatus {
  botName: string;
  status: "success" | "failure" | "creating";
}

const AddBotModal: FC<
  ModalProps & {
    channelID: string;
    statusChange: (s?: AddBotStatus) => void;
  }
> = ({ channelID, statusChange, show, onHide }) => {
  const [botName, setBotName] = useState("");
  const [botPrompt, setBotPrompt] = useState("");
  const apiURL = import.meta.env.DEV
    ? "http://localhost:4000"
    : window.location.protocol + "//" + window.location.host;
  const disableButton = !botName || !botPrompt;

  const addToChannel = async (botID: string) => {
    fetch(
      `${apiURL}/chat/provider/localchat/channels/${channelID}/bots/${botID}`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      },
    )
      .then((resp) => {
        if (resp.ok) {
          if (statusChange) statusChange();
        } else {
          if (statusChange)
            statusChange({ botName: botName, status: "failure" });
        }
      })
      .catch(() => {
        if (statusChange) statusChange({ botName: botName, status: "failure" });
      });
  };
  const createBot = async () => {
    if (statusChange) statusChange({ botName: botName, status: "creating" });
    onHide();

    fetch(`${apiURL}/bots`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        name: botName,
        prompt: botPrompt,
        llm: "openai",
      }),
    })
      .then((resp) => {
        if (resp.ok) {
          resp.json().then((data) => {
            addToChannel(data.ID);
          });
        } else {
          if (statusChange)
            statusChange({ botName: botName, status: "failure" });
        }
      })
      .catch(() => {
        if (statusChange) statusChange({ botName: botName, status: "failure" });
      });
  };

  return (
    <Modal show={show} onHide={onHide}>
      <div className="text-white">
        <DialogTitle className="font-bold text-xl mb-4">
          Create a Bot
        </DialogTitle>

        <label className="flex flex-col">
          <span className="text-gray-400 text-sm font-semibold leading-6">
            Name
          </span>
          <input
            type="text"
            className="w-full rounded-sm border-gray-500 text-black placeholder-black/40 focus:ring-0 focus:border-gray-500"
            placeholder="Adam"
            value={botName}
            onChange={(e) => setBotName(e.target.value)}
          />
        </label>

        <label className="flex flex-col mt-4">
          <span className="text-gray-400 text-sm font-semibold leading-6">
            Bot Description
          </span>
          <textarea
            rows={3}
            className="w-full rounded-sm border-gray-500 text-black placeholder-black/40 focus:ring-0 focus:border-gray-500"
            placeholder="A depressed accountant"
            value={botPrompt}
            onChange={(e) => setBotPrompt(e.target.value)}
          />
        </label>

        <div className="flex space-x-4 justify-end mt-6">
          <Button
            size="sm"
            mode="light"
            disabled={disableButton}
            onClick={createBot}
          >
            Create
          </Button>
        </div>
      </div>
    </Modal>
  );
};

export default AddBotModal;
