
use structsy::{Structsy, StructsyError, StructsyTx};
use structsy_derive::{Persistent, queries};

#[derive(Persistent)]
struct DirectoryNode {
    dirname: string,
    path: string,
    files: Vec<FileNode>
}

#[derive(PersistentEmbedded)]
struct FileNode {
    filename: string,
    path: string,
    hash: string,
    combined_hash: string,
}
#[queries(MyData)]
trait FileNode {
    fn search(self, name:&str) -> Self;
}

pub fn open_db(path: string) -> Result<Structsy, StructsyError> {
    let db = Structsy::open(path)?;
    db.define::<DirectoryNode>()?
}


pub fn add(dir: DirectoryNode) -> Result<(), StructsyError> {
    let db = open_db("my_db.db")?;
    let mut tx = db.begin()?;
    tx.insert(&dir)?;
    tx.commit()?;
    Ok(())
}