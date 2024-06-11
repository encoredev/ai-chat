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
        <div>
          <span
            className="text-xs whitespace-nowrap border border-white/20 rounded-md mt-2 cursor-pointer px-2 py-1"
            onClick={(event) => selectContents(event.target)}
          >
            {url}
          </span>
        </div>
      </div>
    </Modal>
  );
};

export default InviteFriendModal;

const selectContents = (el: any) => {
  let range = document.createRange();
  range.selectNodeContents(el);
  let sel = window.getSelection()!;
  sel.removeAllRanges();
  sel.addRange(range);
};
