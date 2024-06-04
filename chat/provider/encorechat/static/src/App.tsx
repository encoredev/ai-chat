import "bootstrap/dist/css/bootstrap.min.css";
import './App.css';
import {BrowserRouter, Routes, Route, useSearchParams} from "react-router-dom";
import "@chatscope/chat-ui-kit-styles/dist/default/styles.min.css";
import {nanoid} from "nanoid";
import React from "react";
import {EncoreChat} from "./EncoreChat";
import {Home} from "./Home";

function App() {
  return (
    <BrowserRouter basename="/encorechat/demo">
      <EncoreRoutes/>
    </BrowserRouter>
  );
}

function EncoreRoutes() {
  const [searchParams, setSearchParams] = useSearchParams();
  return (
    <Routes>
      <Route path="/">
        <Route index element={<Home />} />
        <Route path="/chat" element={<EncoreChat channelID={searchParams.get("channel") || nanoid()} userName={searchParams.get("name") || "Sam"} />} />
      </Route>
    </Routes>
  );
}

export default App;