import {getAxiosInstance} from "../Auth";

export const ConvertBase64 = (file) => {
    return new Promise((resolve, reject) => {
        const fileReader = new FileReader();
        fileReader.readAsBinaryString(file)
        fileReader.onload = () => {
            resolve(fileReader.result);
        }
        fileReader.onerror = (error) => {
            reject(error);
        }
    })
}

export const ConvertImageToString = async (file) => {
    const blobString = await ConvertBase64(file);
    return btoa(blobString);
}

export function HumanFileSize(bytes, si=false, dp=1) {
    const thresh = si ? 1000 : 1024;

    if (Math.abs(bytes) < thresh) {
        return bytes + ' B';
    }

    const units = si
        ? ['kB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
        : ['KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB', 'ZiB', 'YiB'];
    let u = -1;
    const r = 10**dp;

    do {
        bytes /= thresh;
        ++u;
    } while (Math.round(Math.abs(bytes) * r) / r >= thresh && u < units.length - 1);


    return bytes.toFixed(dp) + '' + units[u];
}

export async function FilecoinPrice() {
    const url = "https://api.coingecko.com/api/v3/simple/price?ids=filecoin&vs_currencies=usd";
    const instance = getAxiosInstance();

    const result = await instance.get(
        url
    );

    return result.data.filecoin.usd;
}

export async function FiatPrice(amount) {
    const filecoinPrice = await FilecoinPrice();
    var formatter = new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: 'USD',
        maximumFractionDigits: 4,
    });

    return formatter.format(filecoinPrice*amount);
}