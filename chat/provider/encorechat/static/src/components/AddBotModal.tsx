import Button from 'react-bootstrap/Button';
import Modal from 'react-bootstrap/Modal';
import {Form} from "react-bootstrap";
import {useState} from "react";

export function AddBotModal(props:any) {
  const [botName, setBotName] = useState("");
  const [botPrompt, setBotPrompt] = useState("");
  const channelID = props.channelID;
  const apiURL = window.location.port === "3000" ? "http://localhost:4000" : window.location.protocol + "//" + window.location.host;

  const addToChannel = async (botID: string) => {
    fetch(`${apiURL}/chat/provider/encorechat/channels/${channelID}/bots/${botID}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
    })
  }
  const createBot = async () => {
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
      }).then((resp) => {
        if (resp.ok) {
          resp.json().then((data) => {
            addToChannel(data.ID)
          })
        }
      })
      props.onHide();
  }

  return (
    <Modal
      {...props}
      size="lg"
      aria-labelledby="contained-modal-title-vcenter"
      centered
    >
      <Modal.Header closeButton>
        <Modal.Title id="contained-modal-title-vcenter">
          Create a Bot
        </Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <Form>
          <Form.Group className="mb-3" controlId="exampleForm.ControlInput1">
            <Form.Label>Bot Name</Form.Label>
            <Form.Control type="name" placeholder="Adam" value={botName} onChange={e => setBotName(e.target.value)} />
          </Form.Group>
          <Form.Group className="mb-3" controlId="exampleForm.ControlTextarea1">
            <Form.Label>Bot Description</Form.Label>
            <Form.Control as="textarea" rows={3} placeholder="A depressed accountant" value={botPrompt} onChange={e => setBotPrompt(e.target.value)  } />
          </Form.Group>
        </Form>
      </Modal.Body>
      <Modal.Footer>
        <Button onClick={props.onHide}>Cancel</Button>
        <Button onClick={createBot}>Create</Button>
      </Modal.Footer>
    </Modal>
  );
}