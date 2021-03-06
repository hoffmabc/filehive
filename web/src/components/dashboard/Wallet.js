import React, {useEffect, useState} from 'react';
import { Link } from 'react-router-dom';
import axios from "axios";
import QRCode from 'qrcode.react';
import ErrorBox, {SuccessBox} from "../ErrorBox";
import {getAxiosInstance} from "../Auth";
import { FilecoinPrice } from "../utilities/images";

export const GetWalletBalance = async () => {

    const csrftoken = localStorage.getItem('csrf_token');
    const instance = axios.create({
        baseURL: "",
        headers: { "x-csrf-token": csrftoken }
    })

    const loginUrl = "/api/v1/wallet/balance";
    const apiReq = await instance.get(
        loginUrl
    );
    console.debug(apiReq);

    return apiReq.data.Balance;
}

export const GetWalletAddress = async () => {

    const csrftoken = localStorage.getItem('csrf_token');
    const instance = axios.create({
        baseURL: "",
        headers: { "x-csrf-token": csrftoken }
    })

    const loginUrl = "/api/v1/wallet/address";
    const apiReq = await instance.get(
        loginUrl
    );
    console.debug(apiReq);

    return apiReq.data.Address;
}

export default function Wallet() {

    const [balance, setBalance] = useState(0);
    const [address, setAddress] = useState("");
    const [filecoinAddress, setFilecoinAddress] = useState("");
    const [amount, setAmount] = useState("");
    const [txAmount, setTxAmount] = useState(0);
    const [recipient, setRecipient] = useState("");
    const [error, setError] = useState("");
    const [success, setSuccess] = useState("");
    const [filecoinPrice, setFilecoinPrice] = useState("");

    const qrSettings = {
        width: 188,
        height: 188
    }

    useEffect(() => {
        const fetchData = async() => {
            const balance = await GetWalletBalance();
            setBalance(balance);
            const address = await GetWalletAddress();
            setAddress(address);
            setFilecoinAddress(address);

            const filecoinPrice = await FilecoinPrice();

            var formatter = new Intl.NumberFormat('en-US', {
                style: 'currency',
                currency: 'USD',
            });

            const balanceUSD = formatter.format(filecoinPrice*balance);

            setFilecoinPrice(balanceUSD);
        };
        fetchData();
    }, []);

    const HandleSendSubmit = async (e) => {
        e.preventDefault();

        // Validate form
        if(amount <= 0) {
            setError("Please enter an amount to send");
            return false;
        }
        if(address === "") {
            setError("Please enter a properly formatted FIL address");
            return false;
        }

        const data = { amount: parseFloat(amount), address: recipient };

        const sendCoins = async () => {

            const sendUrl = "/api/v1/wallet/send";

            try {

                const instance = getAxiosInstance();

                await instance.post(
                    sendUrl,
                    data
                ).then((data)=>{
                    setSuccess("Funds sent successfully");
                    setError("");
                    const fetchData = async() => {
                        const balance = await GetWalletBalance();
                        setBalance(balance);
                        setAmount("");
                        setRecipient("");
                    };
                    fetchData();
                }).catch(error => {
                    console.log("Send Failure", error.response);
                    setSuccess("");
                    setError(error.response.data.error);
                });
                return false;

            } catch(err) {

            }
        };
        sendCoins();
        // setRecipient("");
        // setAmount(0);

    }

    return (

        <div className="maincontent margins-30">
            <h2>Wallet <span className="h2-subtitle">{balance} ({filecoinPrice})</span></h2>

            <div className="withdrawal-deposit-container">
                <div className="wd-container">
                    <h3>Deposit</h3>
                    <div className="wd-description">Send FIL to the address below to add funds to your wallet.</div>
                    <div className="qr-code-deposit">
                        <QRCode value={filecoinAddress} size="99" imageSettings={qrSettings} />
                    </div>
                    <div className="center">{address}</div>
                    <div className="center"><a className="orange-link" onClick={() =>  navigator.clipboard.writeText(address)}>copy</a></div>
                </div>
                <div className="wd-container form-540">
                    <h3>Withdrawal</h3>
                    <p>Specify a FIL address below to send your funds to.</p>
                    <form onSubmit={HandleSendSubmit}>
                        <label>
                            Amount*
                            <input type="text" placeholder="0" value={amount}
                                   onChange={e => setAmount(e.target.value)}/>
                        </label>
                        <label>
                            To (FIL address)*
                            <input type="text" value={recipient} placeholder="e.g. f1cadxk4yywa7hfaiz3rs23t3wmyn7cjcdy5rtm4q"
                                   onChange={e => setRecipient(e.target.value)}/>
                        </label>
                        <div>
                            <input type="submit" value="Send" className="orange-button"/>
                        </div>
                        <div className="note">
                            <p className="mini-light-description">*Note that all transactions incur a gas fee.</p>
                        </div>

                        {error &&
                        <ErrorBox message={error}/>
                        }
                        {success &&
                        <SuccessBox message={success}/>
                        }

                    </form>
                </div>
            </div>
            {/*<div className="transaction-ledger">*/}
            {/*    <h3>Transactions</h3>*/}

            {/*</div>*/}


        </div>

    );
}