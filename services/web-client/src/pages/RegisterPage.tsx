import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { register } from "../api/client";

export default function RegisterPage() {
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const res = await register(username, email, password);
      localStorage.setItem("token", res.token);
      localStorage.setItem("user_id", res.user_id);
      navigate("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "registration failed");
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center">
      <form onSubmit={handleSubmit} className="flex flex-col gap-4 w-80">
        <h1 className="text-2xl font-bold">Create Account</h1>
        {error && <p className="text-red-400 text-sm">{error}</p>}
        <input
          className="bg-neutral-800 rounded px-3 py-2"
          placeholder="Username"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
        />
        <input
          className="bg-neutral-800 rounded px-3 py-2"
          type="email"
          placeholder="Email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
        />
        <input
          className="bg-neutral-800 rounded px-3 py-2"
          type="password"
          placeholder="Password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <button className="bg-blue-600 rounded py-2 font-medium" type="submit">
          Create Account
        </button>
        <a href="/login" className="text-sm text-neutral-400 text-center">
          Already have an account?
        </a>
      </form>
    </div>
  );
}
