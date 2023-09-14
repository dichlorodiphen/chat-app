import './Auth.css'

import Logo from '../components/Logo';
import { useState } from 'react';

type AuthProps = {
    setToken: React.Dispatch<React.SetStateAction<string>>
};

function Auth({ setToken }: AuthProps) {
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");

    function onUsernameChange(e: React.FormEvent<HTMLInputElement>) {
        setUsername(e.currentTarget.value);
    }

    function onPasswordChange(e: React.FormEvent<HTMLInputElement>) {
        setPassword(e.currentTarget.value);
    }

    async function signUp() {
        console.log(`Signing up with username ${username} and password ${password}`);

        const data = await fetch("http://127.0.0.1:8000/users/signup", {
            method: "POST",
            body: JSON.stringify({
                "username": username,
                "password": password,
            }),
        });

        const token = await data.text();
        console.log(`Got the following from signup endpoint: ${token}`);
        if (data.status == 201) {
            setToken(token);
        }
    }

    async function logIn() {
        console.log(`Trying to log in with username ${username} and password ${password}`);

        const data = await fetch("http://127.0.0.1:8000/users/login", {
            method: "POST",
            body: JSON.stringify({
                "username": username,
                "password": password,
            }),
        });

        const token = await data.text();
        console.log(`Got the following from login endpoint: ${token}`);
        if (data.status == 200) {
            setToken(token);
        }
    }

    return (
        <div>
            <Logo></Logo>

            <div className='auth'>
                <p>Sign up or log in to get started!</p>
                <form>
                    <div className='text-input'>
                        <label>Username</label><br />
                        <input value={username} onChange={onUsernameChange} type="text"></input><br />
                    </div>

                    <div className='text-input'>
                        <label>Password</label><br />
                        <input value={password} onChange={onPasswordChange} type="password"></input><br />
                    </div>
                </form>
                <div className='auth-buttons'>
                    <input type="submit" onClick={logIn} value="Log in"></input>
                    <input type="submit" onClick={signUp} value="Sign up"></input>
                </div>
            </div>
        </div>
    );
}

export default Auth;