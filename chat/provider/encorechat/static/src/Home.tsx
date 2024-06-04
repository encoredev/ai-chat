
import {useNavigate, createSearchParams, useSearchParams} from "react-router-dom";
import {useState, FormEvent} from "react";
import {nanoid} from "nanoid";

export const Home = () => {
  let [username, setUsername] = useState("Sam");
  const [status, setStatus] = useState('typing');
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setStatus('submitting');
    navigate({
      pathname: "/chat",
      search: createSearchParams({name: username, channel:searchParams.get("channel") || nanoid() }).toString()
    });
  }

  return (
    <div>
      <h1>Home</h1>
      <p>Welcome to the home page</p>
      <form onSubmit={handleSubmit}>
        <input type="text"
               placeholder="Enter your username"
               value={username}
               onChange={(event) => setUsername(event.target.value)}
        />
        <button disabled={
          username.length === 0 ||
          status === 'submitting'
        }>
          Submit
        </button>
      </form>
    </div>
)
}
