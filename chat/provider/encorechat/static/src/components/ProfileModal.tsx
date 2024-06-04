import Button from 'react-bootstrap/Button';
import Modal from 'react-bootstrap/Modal';
import {Col, Row, Container, Image} from "react-bootstrap";
import {User} from "@chatscope/use-chat";

export function ProfileModal(props:any) {
  if(props.user === undefined){
    return (<></>);
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
          {props.user.username}
        </Modal.Title>
      </Modal.Header>
      <Modal.Body >
        <Container>
          <Row>
            <Col md="auto">
              <Image src={props.user.avatar} width="250" height="250" rounded />
            </Col>
            <Col>
              {props.user.bio}
            </Col>
          </Row>
        </Container>
      </Modal.Body>
      <Modal.Footer>
        <Button onClick={props.onHide}>Close</Button>
      </Modal.Footer>
    </Modal>
  );
}