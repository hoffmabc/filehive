import React from 'react'
import {useHistory} from "react-router-dom";
import {HumanFileSize} from "./utilities/images";
import TimeAgo from 'javascript-time-ago'
import en from 'javascript-time-ago/locale/en'

TimeAgo.addDefaultLocale(en);

function DataSetsRow(props) {
    const history = useHistory();

    const timeAgo = new TimeAgo('en-US')

    const imageFilename = props.metadata.imageFilename;
    const title = props.metadata.title;
    const shortDescription = props.metadata.shortDescription;
    const price = Number.parseFloat(props.metadata.price).toFixed(8).toString().replace(/\.?0+$/,"");
    const fileType = props.metadata.fileType;
    const fileSize = HumanFileSize(props.metadata.fileSize, true);
    const username = props.metadata.username;
    const timestamp = timeAgo.format(Date.parse(props.metadata.createdAt));

    const buttonText = (props.rowType === "edit") ? "Edit" : "Details";
    const gotoPage = (props.rowType === "edit") ? '/dataset/'+props.metadata.id+'/edit' : '/dataset/'+props.metadata.id;

    const datasetImage = "/api/v1/image/" + imageFilename;

    const handleClickDatasetRow = (e) => {
        history.push(gotoPage);
    }

    return (
        <div class="datasets-row" onClick={handleClickDatasetRow}>
            <div className="datasets-row-image">
                <img className="datasets-image" src={datasetImage}/>
            </div>
            <div className="datasets-row-info">
                <div className="mini-bold-title">{title}</div>
                <div className="mini-description">{shortDescription}</div>
                <div className="mini-light-description tag-container">
                    <div>{fileType}</div>
                    <div>{fileSize}</div>
                    <div>{timestamp}</div>
                    <div>{username}</div>
                </div>
            </div>
            <div className="datasets-details">
                <div><button className="normal-button">{buttonText}</button></div>
                <div className="small-orange-text">{price} FIL</div>
            </div>
        </div>
    )
}

export default DataSetsRow