import { FC } from "react";
import { DialogTitle } from "@headlessui/react";
import Modal, { ModalProps } from "./Modal.tsx";

const InviteFriendModal: FC<
  ModalProps & {
    channelID: string;
  }
> = ({ channelID, show, onHide }) => {
  const url = location.origin + `?channel=${channelID}`;
  return (
    <Modal show={show} onHide={onHide}>
      <div className="text-white">
        <DialogTitle className="font-bold text-xl mb-4">
          Invite a friend
        </DialogTitle>

        <p className="mb-4 text-sm text-gray-400">
          Send this URL to a friend to let them join the chat:
        </p>
        <div className="overflow-x-auto overflow-y-hidden border border-white/20 rounded-md mt-2 w-fit max-w-full">
          <span className="text-xs whitespace-nowrap px-2 py-1">{url}</span>
        </div>
      </div>
    </Modal>
  );
};

export default InviteFriendModal;
