import { FC, useEffect, useState } from "react";
import { User } from "@chatscope/use-chat";
import { DialogTitle } from "@headlessui/react";
import Modal, { ModalProps } from "./Modal.tsx";

const ProfileModal: FC<
  ModalProps & {
    user?: User;
  }
> = ({ user, show, onHide }) => {
  const [currentUser, setCurrentUser] = useState<User | undefined>(user);

  useEffect(() => {
    if (user) setCurrentUser(user);
  }, [user]);

  return (
    <Modal show={show} onHide={onHide}>
      <div>
        <img src={currentUser?.avatar} className="w-full rounded-md" />
        <div className="mt-3 sm:mt-5">
          <DialogTitle
            as="h3"
            className="flex items-center text-black text-2xl font-semibold leading-6"
          >
            {currentUser?.username}
            <span className="ml-2.5 inline-block h-2 w-2 flex-shrink-0 rounded-full bg-green">
              <span className="sr-only">Online</span>
            </span>
          </DialogTitle>
          <div className="mt-2 text-gray-500">
            <h4 className="font-semibold">Bio</h4>
            <p className="text-sm">{currentUser?.bio}</p>
          </div>
        </div>
      </div>
    </Modal>
  );
};

export default ProfileModal;
