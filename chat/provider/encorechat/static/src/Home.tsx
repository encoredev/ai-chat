
import {useNavigate, createSearchParams, useSearchParams} from "react-router-dom";
import {useState, FormEvent} from "react";
import {nanoid} from "nanoid";
import logo from "./assets/aichat.png"
import {Card, Form} from "react-bootstrap";
import Button from "react-bootstrap/Button";

export const Home = () => {
  let [username, setUsername] = useState("Sam");
  const [status, setStatus] = useState('typing');
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  async function joinChat() {
    setStatus('submitting')
    navigate({
      pathname: "/chat",
      search: createSearchParams({name: username, channel:searchParams.get("channel") || nanoid() }).toString()
    });
  }

  return (
    <Card style={{width: '18rem'}}>
      <Card.Img variant="top" src={logo}/>
      <Card.Body>
        <Form.Group className="mb-3" controlId="exampleForm.ControlInput1">
          <Form.Control type="text" placeholder="Your name" value={username}
                        onChange={e => setUsername(e.target.value)}/>
        </Form.Group>
        <Button variant="primary" onClick={joinChat} disabled={
          username.length === 0 ||
          status === 'submitting'
        }>Join Chat</Button>
      </Card.Body>
    </Card>
  )
}
